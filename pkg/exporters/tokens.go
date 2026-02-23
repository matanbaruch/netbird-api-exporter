package exporters

import (
	"context"
	"os"
	"time"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// TokensExporter handles tokens-specific metrics collection
// NOTE: This exporter is opt-in due to potential high cardinality
type TokensExporter struct {
	client  *nbclient.Client
	enabled bool

	// Prometheus metrics for tokens
	tokensTotal       *prometheus.GaugeVec
	tokensByUser      *prometheus.GaugeVec
	tokenInfo         *prometheus.GaugeVec
	tokenExpiresAt    *prometheus.GaugeVec
	tokenLastUsedAt   *prometheus.GaugeVec
	scrapeErrorsTotal *prometheus.CounterVec
	scrapeDuration    *prometheus.HistogramVec
}

// NewTokensExporter creates a new tokens exporter
// This exporter is disabled by default and must be explicitly enabled via ENABLE_TOKENS_EXPORTER=true
func NewTokensExporter(client *nbclient.Client) *TokensExporter {
	enabled := os.Getenv("ENABLE_TOKENS_EXPORTER") == "true"

	if enabled {
		logrus.Info("Tokens exporter is enabled (high cardinality metrics)")
	} else {
		logrus.Info("Tokens exporter is disabled (set ENABLE_TOKENS_EXPORTER=true to enable)")
	}

	return &TokensExporter{
		client:  client,
		enabled: enabled,

		tokensTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_tokens",
				Help: "Total number of NetBird personal access tokens across all users",
			},
			[]string{},
		),

		tokensByUser: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_tokens_by_user",
				Help: "Number of personal access tokens per user",
			},
			[]string{"user_id"},
		),

		tokenInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_token_info",
				Help: "Information about NetBird personal access tokens (always 1)",
			},
			[]string{"token_id", "token_name", "user_id"},
		),

		tokenExpiresAt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_token_expires_at_timestamp",
				Help: "Unix timestamp when the personal access token expires",
			},
			[]string{"token_id", "token_name", "user_id"},
		),

		tokenLastUsedAt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_token_last_used_timestamp",
				Help: "Unix timestamp when the personal access token was last used",
			},
			[]string{"token_id", "token_name", "user_id"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_tokens_scrape_errors_total",
				Help: "Total number of errors encountered while scraping tokens",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_tokens_scrape_duration_seconds",
				Help: "Time spent scraping tokens from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *TokensExporter) Describe(ch chan<- *prometheus.Desc) {
	if !e.enabled {
		return
	}

	e.tokensTotal.Describe(ch)
	e.tokensByUser.Describe(ch)
	e.tokenInfo.Describe(ch)
	e.tokenExpiresAt.Describe(ch)
	e.tokenLastUsedAt.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *TokensExporter) Collect(ch chan<- prometheus.Metric) {
	if !e.enabled {
		return
	}

	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.tokensTotal.Reset()
	e.tokensByUser.Reset()
	e.tokenInfo.Reset()
	e.tokenExpiresAt.Reset()
	e.tokenLastUsedAt.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// First, get all users to iterate through their tokens
	users, err := e.client.Users.List(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch users for tokens")
		e.scrapeErrorsTotal.WithLabelValues("fetch_users").Inc()
		return
	}

	e.updateMetrics(ctx, users)

	// Collect all metrics
	e.tokensTotal.Collect(ch)
	e.tokensByUser.Collect(ch)
	e.tokenInfo.Collect(ch)
	e.tokenExpiresAt.Collect(ch)
	e.tokenLastUsedAt.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on tokens data
func (e *TokensExporter) updateMetrics(ctx context.Context, users []api.User) {
	totalTokens := 0

	for _, user := range users {
		tokens, err := e.client.Tokens.List(ctx, user.Id)
		if err != nil {
			logrus.WithError(err).WithField("user_id", user.Id).Debug("Failed to fetch tokens for user")
			e.scrapeErrorsTotal.WithLabelValues("fetch_user_tokens").Inc()
			continue
		}

		// Set tokens per user
		e.tokensByUser.WithLabelValues(user.Id).Set(float64(len(tokens)))
		totalTokens += len(tokens)

		for _, token := range tokens {
			// Set token info metric
			infoLabels := []string{token.Id, token.Name, user.Id}
			e.tokenInfo.WithLabelValues(infoLabels...).Set(1)

			// Set expiration timestamp
			e.tokenExpiresAt.WithLabelValues(infoLabels...).Set(float64(token.ExpirationDate.Unix()))

			// Set last used timestamp if available
			if token.LastUsed != nil && !token.LastUsed.IsZero() {
				e.tokenLastUsedAt.WithLabelValues(infoLabels...).Set(float64(token.LastUsed.Unix()))
			}
		}
	}

	// Set total tokens
	e.tokensTotal.WithLabelValues().Set(float64(totalTokens))

	logrus.WithFields(logrus.Fields{
		"total_tokens": totalTokens,
		"total_users":  len(users),
	}).Debug("Updated tokens metrics")
}
