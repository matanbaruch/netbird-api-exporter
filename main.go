package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/matanbaruch/netbird-api-exporter/pkg/exporters"
	"github.com/matanbaruch/netbird-api-exporter/pkg/utils"
)

// securityHeadersMiddleware adds security headers to all HTTP responses
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS protection in older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent caching of sensitive endpoints
		if r.URL.Path == "/metrics" || r.URL.Path == "/health" {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		next.ServeHTTP(w, r)
	})
}

// debugLoggingMiddleware logs HTTP requests when debug level is enabled
func debugLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logrus.WithFields(logrus.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Debug("HTTP request received")

		// Create a custom ResponseWriter to capture status code
		ww := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		logrus.WithFields(logrus.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": ww.statusCode,
			"duration":    duration,
		}).Debug("HTTP request completed")
	})
}

// wrappedWriter wraps http.ResponseWriter to capture status code
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func main() {
	// Configuration from environment variables
	netbirdURL := utils.GetEnvWithDefault("NETBIRD_API_URL", "https://api.netbird.io")
	netbirdToken := os.Getenv("NETBIRD_API_TOKEN")
	listenAddr := utils.GetEnvWithDefault("LISTEN_ADDRESS", ":8080")
	metricsPath := utils.GetEnvWithDefault("METRICS_PATH", "/metrics")
	logLevel := utils.GetEnvWithDefault("LOG_LEVEL", "info")

	// Check for help flag before validating token
	helpFlag := false
	for _, arg := range os.Args {
		if arg == "--help" || arg == "-h" {
			helpFlag = true
			break
		}
	}

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// If help flag is present, print default help message and exit
	// This is a simple way to handle help without fully parsing flags if token is missing
	if helpFlag {
		// This will print default help, but since we don't use `flag` package for all vars,
		// it might not be comprehensive. A more robust solution would involve `flag.Usage`.
		// For now, let's assume the user knows about env vars from README.
		fmt.Fprintf(os.Stderr, "Usage of %s:\\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  This application is configured primarily via environment variables.\\n")
		fmt.Fprintf(os.Stderr, "  Key environment variables:\\n")
		fmt.Fprintf(os.Stderr, "    NETBIRD_API_URL: NetBird API endpoint (default: https://api.netbird.io)\\n")
		fmt.Fprintf(os.Stderr, "    NETBIRD_API_TOKEN: NetBird API token (required)\\n")
		fmt.Fprintf(os.Stderr, "    LISTEN_ADDRESS: HTTP server listen address (default: :8080)\\n")
		fmt.Fprintf(os.Stderr, "    METRICS_PATH: Metrics endpoint path (default: /metrics)\\n")
		fmt.Fprintf(os.Stderr, "    LOG_LEVEL: Logging level (default: info)\\n")
		fmt.Fprintf(os.Stderr, "  Use --help or -h to display this message.\\n")
		os.Exit(0)
	}

	// Validate required configuration
	if netbirdToken == "" {
		logrus.Fatal("NETBIRD_API_TOKEN environment variable is required")
	}

	logrus.WithFields(logrus.Fields{
		"netbird_url":  netbirdURL,
		"listen_addr":  listenAddr,
		"metrics_path": metricsPath,
		"log_level":    logLevel,
	}).Info("Starting NetBird API Exporter")

	// Create exporter
	exporter := exporters.NewNetBirdExporter(netbirdURL, netbirdToken)

	// Register exporter
	prometheus.MustRegister(exporter)

	// Create HTTP server
	mux := http.NewServeMux()

	// Build middleware chain: security headers first, then debug logging
	var handler http.Handler = mux
	handler = securityHeadersMiddleware(handler)
	if logrus.GetLevel() == logrus.DebugLevel {
		handler = debugLoggingMiddleware(handler)
	}

	// Metrics endpoint
	mux.Handle(metricsPath, promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		logrus.Debug("Health check endpoint accessed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339)); err != nil {
			logrus.WithError(err).Error("Failed to write health check response")
		}
	})

	// Root endpoint with information
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Debug("Root endpoint accessed")
		w.Header().Set("Content-Type", "text/html")
		if _, err := fmt.Fprintf(w, `
		<html>
		<head><title>NetBird API Exporter</title></head>
		<body>
		<h1>NetBird API Exporter</h1>
		<p>This is a Prometheus exporter for NetBird API metrics.</p>
		<ul>
		<li><a href="%s">Metrics</a></li>
		<li><a href="/health">Health Check</a></li>
		</ul>
		<h2>Available Metrics</h2>
		<ul>
		<li><strong>Peers API:</strong> Connection status, OS distribution, geographic distribution, SSH status, login status</li>
		<li><strong>Groups API:</strong> Group counts and membership</li>
		<li><strong>Users API:</strong> User counts, roles, status, permissions</li>
		<li><strong>DNS API:</strong> Nameserver groups, DNS settings, nameserver configurations</li>
		<li><strong>Networks API:</strong> Network counts, routers, resources, policies, routing peers</li>
		</ul>
		</body>
		</html>
		`, metricsPath); err != nil {
			logrus.WithError(err).Error("Failed to write root page response")
		}
	})

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logrus.Info("Shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logrus.WithError(err).Error("Error during server shutdown")
		}
		cancel()
	}()

	// Start server
	logrus.WithField("address", listenAddr).Info("Starting HTTP server")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.WithError(err).Fatal("HTTP server error")
	}

	<-ctx.Done()
	logrus.Info("Server stopped")
}
