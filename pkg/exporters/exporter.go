package exporters

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
)

// sanitizePanicValue returns a safe string representation of a panic value
// without exposing sensitive information
func sanitizePanicValue(r interface{}) string {
	if r == nil {
		return "unknown panic"
	}

	// Convert to string but truncate to prevent information leakage
	panicStr := fmt.Sprintf("%v", r)

	// Limit length to prevent excessive logging
	maxLen := 200
	if len(panicStr) > maxLen {
		panicStr = panicStr[:maxLen] + "... (truncated)"
	}

	return panicStr
}

// NetBirdExporter represents the main Prometheus exporter for NetBird APIs
type NetBirdExporter struct {
	client           *nbclient.Client
	peersExporter    *PeersExporter
	groupsExporter   *GroupsExporter
	usersExporter    *UsersExporter
	dnsExporter      *DNSExporter
	networksExporter *NetworksExporter

	// Common metrics
	scrapeDuration prometheus.Histogram
	scrapeErrors   prometheus.Counter
}

// NewNetBirdExporter creates a new NetBird exporter with all sub-exporters
func NewNetBirdExporter(baseURL, token string) *NetBirdExporter {

	client := nbclient.New(baseURL, token)

	return &NetBirdExporter{
		client:           client,
		peersExporter:    NewPeersExporter(client),
		groupsExporter:   NewGroupsExporter(client),
		usersExporter:    NewUsersExporter(client),
		dnsExporter:      NewDNSExporter(client),
		networksExporter: NewNetworksExporter(client),

		scrapeDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name: "netbird_exporter_scrape_duration_seconds",
				Help: "Time spent scraping NetBird API",
			},
		),

		scrapeErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "netbird_exporter_scrape_errors_total",
				Help: "Total number of scrape errors",
			},
		),
	}
}

// Describe implements prometheus.Collector
func (e *NetBirdExporter) Describe(ch chan<- *prometheus.Desc) {
	e.peersExporter.Describe(ch)
	e.groupsExporter.Describe(ch)
	e.usersExporter.Describe(ch)
	e.dnsExporter.Describe(ch)
	e.networksExporter.Describe(ch)
	e.scrapeDuration.Describe(ch)
	e.scrapeErrors.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *NetBirdExporter) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		e.scrapeDuration.Observe(duration.Seconds())
		e.scrapeDuration.Collect(ch)
		e.scrapeErrors.Collect(ch)
		logrus.WithField("total_duration", duration).Debug("Completed NetBird metrics collection")
	}()

	logrus.Debug("Starting NetBird metrics collection")

	// Collect from all sub-exporters
	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", sanitizePanicValue(r)).Error("Panic during peers collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting peers collection")
		e.peersExporter.Collect(ch)
		logrus.Debug("Completed peers collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", sanitizePanicValue(r)).Error("Panic during groups collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting groups collection")
		e.groupsExporter.Collect(ch)
		logrus.Debug("Completed groups collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", sanitizePanicValue(r)).Error("Panic during users collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting users collection")
		e.usersExporter.Collect(ch)
		logrus.Debug("Completed users collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", sanitizePanicValue(r)).Error("Panic during dns collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting DNS collection")
		e.dnsExporter.Collect(ch)
		logrus.Debug("Completed DNS collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", sanitizePanicValue(r)).Error("Panic during networks collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting networks collection")
		e.networksExporter.Collect(ch)
		logrus.Debug("Completed networks collection")
	}()

	// Future exporters can be added here like:
	// e.policiesExporter.Collect(ch)
}
