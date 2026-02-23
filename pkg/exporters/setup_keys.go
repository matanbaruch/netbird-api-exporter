package exporters

import (
	"context"
	"time"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// SetupKeysExporter handles setup keys-specific metrics collection
type SetupKeysExporter struct {
	client *nbclient.Client

	// Prometheus metrics for setup keys
	setupKeysTotal       *prometheus.GaugeVec
	setupKeysByType      *prometheus.GaugeVec
	setupKeysByState     *prometheus.GaugeVec
	setupKeyInfo         *prometheus.GaugeVec
	setupKeyUsedTimes    *prometheus.GaugeVec
	setupKeyUsageLimit   *prometheus.GaugeVec
	setupKeyExpiresAt    *prometheus.GaugeVec
	setupKeyLastUsedAt   *prometheus.GaugeVec
	scrapeErrorsTotal    *prometheus.CounterVec
	scrapeDuration       *prometheus.HistogramVec
}

// NewSetupKeysExporter creates a new setup keys exporter
func NewSetupKeysExporter(client *nbclient.Client) *SetupKeysExporter {
	return &SetupKeysExporter{
		client: client,

		setupKeysTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_keys",
				Help: "Total number of NetBird setup keys",
			},
			[]string{},
		),

		setupKeysByType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_keys_by_type",
				Help: "Number of NetBird setup keys by type (one-off or reusable)",
			},
			[]string{"type"},
		),

		setupKeysByState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_keys_by_state",
				Help: "Number of NetBird setup keys by state (valid, expired, overused, revoked)",
			},
			[]string{"state"},
		),

		setupKeyInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_info",
				Help: "Information about NetBird setup keys (always 1)",
			},
			[]string{"key_id", "key_name", "type", "state", "valid", "revoked", "ephemeral"},
		),

		setupKeyUsedTimes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_used_times",
				Help: "Number of times a setup key has been used",
			},
			[]string{"key_id", "key_name"},
		),

		setupKeyUsageLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_usage_limit",
				Help: "Usage limit for a setup key (0 = unlimited)",
			},
			[]string{"key_id", "key_name"},
		),

		setupKeyExpiresAt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_expires_at_timestamp",
				Help: "Unix timestamp when the setup key expires",
			},
			[]string{"key_id", "key_name"},
		),

		setupKeyLastUsedAt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_last_used_timestamp",
				Help: "Unix timestamp when the setup key was last used",
			},
			[]string{"key_id", "key_name"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_setup_keys_scrape_errors_total",
				Help: "Total number of errors encountered while scraping setup keys",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_setup_keys_scrape_duration_seconds",
				Help: "Time spent scraping setup keys from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *SetupKeysExporter) Describe(ch chan<- *prometheus.Desc) {
	e.setupKeysTotal.Describe(ch)
	e.setupKeysByType.Describe(ch)
	e.setupKeysByState.Describe(ch)
	e.setupKeyInfo.Describe(ch)
	e.setupKeyUsedTimes.Describe(ch)
	e.setupKeyUsageLimit.Describe(ch)
	e.setupKeyExpiresAt.Describe(ch)
	e.setupKeyLastUsedAt.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *SetupKeysExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.setupKeysTotal.Reset()
	e.setupKeysByType.Reset()
	e.setupKeysByState.Reset()
	e.setupKeyInfo.Reset()
	e.setupKeyUsedTimes.Reset()
	e.setupKeyUsageLimit.Reset()
	e.setupKeyExpiresAt.Reset()
	e.setupKeyLastUsedAt.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	setupKeys, err := e.client.SetupKeys.List(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch setup keys")
		e.scrapeErrorsTotal.WithLabelValues("fetch_setup_keys").Inc()
		return
	}

	e.updateMetrics(setupKeys)

	// Collect all metrics
	e.setupKeysTotal.Collect(ch)
	e.setupKeysByType.Collect(ch)
	e.setupKeysByState.Collect(ch)
	e.setupKeyInfo.Collect(ch)
	e.setupKeyUsedTimes.Collect(ch)
	e.setupKeyUsageLimit.Collect(ch)
	e.setupKeyExpiresAt.Collect(ch)
	e.setupKeyLastUsedAt.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on setup keys data
func (e *SetupKeysExporter) updateMetrics(setupKeys []api.SetupKey) {
	typeCount := make(map[string]int)
	stateCount := make(map[string]int)

	for _, key := range setupKeys {
		// Count by type and state
		typeCount[key.Type]++
		stateCount[key.State]++

		// Convert booleans to strings
		validStr := "false"
		if key.Valid {
			validStr = "true"
		}
		revokedStr := "false"
		if key.Revoked {
			revokedStr = "true"
		}
		ephemeralStr := "false"
		if key.Ephemeral {
			ephemeralStr = "true"
		}

		// Set key info metric
		infoLabels := []string{key.Id, key.Name, key.Type, key.State, validStr, revokedStr, ephemeralStr}
		e.setupKeyInfo.WithLabelValues(infoLabels...).Set(1)

		// Set usage metrics
		usageLabels := []string{key.Id, key.Name}
		e.setupKeyUsedTimes.WithLabelValues(usageLabels...).Set(float64(key.UsedTimes))
		e.setupKeyUsageLimit.WithLabelValues(usageLabels...).Set(float64(key.UsageLimit))

		// Set timestamp metrics
		e.setupKeyExpiresAt.WithLabelValues(usageLabels...).Set(float64(key.Expires.Unix()))
		if !key.LastUsed.IsZero() {
			e.setupKeyLastUsedAt.WithLabelValues(usageLabels...).Set(float64(key.LastUsed.Unix()))
		}
	}

	// Set total and distribution metrics
	e.setupKeysTotal.WithLabelValues().Set(float64(len(setupKeys)))

	for keyType, count := range typeCount {
		e.setupKeysByType.WithLabelValues(keyType).Set(float64(count))
	}

	for state, count := range stateCount {
		e.setupKeysByState.WithLabelValues(state).Set(float64(count))
	}

	logrus.WithFields(logrus.Fields{
		"total_setup_keys": len(setupKeys),
		"by_type":          typeCount,
		"by_state":         stateCount,
	}).Debug("Updated setup keys metrics")
}
