package exporters

import (
	"context"
	"strconv"
	"time"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// SetupKeysExporter handles setup-key-specific metrics collection
type SetupKeysExporter struct {
	client *nbclient.Client

	// Prometheus metrics for setup keys
	setupKeysTotal     *prometheus.GaugeVec
	setupKeysValid     *prometheus.GaugeVec
	setupKeysRevoked   *prometheus.GaugeVec
	setupKeysEphemeral *prometheus.GaugeVec
	setupKeyUsedTimes  *prometheus.GaugeVec
	setupKeyUsageLimit *prometheus.GaugeVec
	setupKeyExpires    *prometheus.GaugeVec
	setupKeyLastUsed   *prometheus.GaugeVec
	setupKeyInfo       *prometheus.GaugeVec
	setupKeyAutoGroups *prometheus.GaugeVec
	scrapeErrorsTotal  *prometheus.CounterVec
	scrapeDuration     *prometheus.HistogramVec
}

// NewSetupKeysExporter creates a new setup keys exporter
func NewSetupKeysExporter(client *nbclient.Client) *SetupKeysExporter {
	return &SetupKeysExporter{
		client: client,

		setupKeysTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_keys",
				Help: "Total number of NetBird setup keys by type and state",
			},
			[]string{"type", "state"},
		),

		setupKeysValid: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_keys_valid",
				Help: "Number of NetBird setup keys grouped by validity status",
			},
			[]string{"valid"},
		),

		setupKeysRevoked: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_keys_revoked",
				Help: "Number of NetBird setup keys grouped by revocation status",
			},
			[]string{"revoked"},
		),

		setupKeysEphemeral: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_keys_ephemeral",
				Help: "Number of NetBird setup keys grouped by ephemeral status",
			},
			[]string{"ephemeral"},
		),

		setupKeyUsedTimes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_used_times",
				Help: "Number of times a NetBird setup key has been used",
			},
			[]string{"key_id", "key_name"},
		),

		setupKeyUsageLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_usage_limit",
				Help: "Usage limit configured for a NetBird setup key (0 means unlimited)",
			},
			[]string{"key_id", "key_name"},
		),

		setupKeyExpires: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_expires_timestamp",
				Help: "Expiration date of a NetBird setup key as a Unix timestamp",
			},
			[]string{"key_id", "key_name"},
		),

		setupKeyLastUsed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_last_used_timestamp",
				Help: "Last usage date of a NetBird setup key as a Unix timestamp",
			},
			[]string{"key_id", "key_name"},
		),

		setupKeyInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_info",
				Help: "Information about NetBird setup keys (always 1)",
			},
			[]string{"key_id", "key_name", "type", "state"},
		),

		setupKeyAutoGroups: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_setup_key_auto_groups_count",
				Help: "Number of auto-assigned groups configured for a NetBird setup key",
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
	e.setupKeysValid.Describe(ch)
	e.setupKeysRevoked.Describe(ch)
	e.setupKeysEphemeral.Describe(ch)
	e.setupKeyUsedTimes.Describe(ch)
	e.setupKeyUsageLimit.Describe(ch)
	e.setupKeyExpires.Describe(ch)
	e.setupKeyLastUsed.Describe(ch)
	e.setupKeyInfo.Describe(ch)
	e.setupKeyAutoGroups.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *SetupKeysExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.setupKeysTotal.Reset()
	e.setupKeysValid.Reset()
	e.setupKeysRevoked.Reset()
	e.setupKeysEphemeral.Reset()
	e.setupKeyUsedTimes.Reset()
	e.setupKeyUsageLimit.Reset()
	e.setupKeyExpires.Reset()
	e.setupKeyLastUsed.Reset()
	e.setupKeyInfo.Reset()
	e.setupKeyAutoGroups.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	e.setupKeysValid.Collect(ch)
	e.setupKeysRevoked.Collect(ch)
	e.setupKeysEphemeral.Collect(ch)
	e.setupKeyUsedTimes.Collect(ch)
	e.setupKeyUsageLimit.Collect(ch)
	e.setupKeyExpires.Collect(ch)
	e.setupKeyLastUsed.Collect(ch)
	e.setupKeyInfo.Collect(ch)
	e.setupKeyAutoGroups.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on setup keys data
func (e *SetupKeysExporter) updateMetrics(setupKeys []api.SetupKey) {
	totalKeys := len(setupKeys)

	typeStateCounts := make(map[string]map[string]int)
	validCounts := make(map[bool]int)
	revokedCounts := make(map[bool]int)
	ephemeralCounts := make(map[bool]int)

	for _, key := range setupKeys {
		if typeStateCounts[key.Type] == nil {
			typeStateCounts[key.Type] = make(map[string]int)
		}
		typeStateCounts[key.Type][key.State]++

		validCounts[key.Valid]++
		revokedCounts[key.Revoked]++
		ephemeralCounts[key.Ephemeral]++

		keyLabels := []string{key.Id, key.Name}

		e.setupKeyUsedTimes.WithLabelValues(keyLabels...).Set(float64(key.UsedTimes))
		e.setupKeyUsageLimit.WithLabelValues(keyLabels...).Set(float64(key.UsageLimit))
		if !key.Expires.IsZero() {
			e.setupKeyExpires.WithLabelValues(keyLabels...).Set(float64(key.Expires.Unix()))
		}
		if !key.LastUsed.IsZero() {
			e.setupKeyLastUsed.WithLabelValues(keyLabels...).Set(float64(key.LastUsed.Unix()))
		}
		e.setupKeyAutoGroups.WithLabelValues(keyLabels...).Set(float64(len(key.AutoGroups)))
		e.setupKeyInfo.WithLabelValues(key.Id, key.Name, key.Type, key.State).Set(1)
	}

	// Set counts by type and state
	for keyType, stateCounts := range typeStateCounts {
		for state, count := range stateCounts {
			e.setupKeysTotal.WithLabelValues(keyType, state).Set(float64(count))
		}
	}

	// Set validity metrics
	for valid, count := range validCounts {
		e.setupKeysValid.WithLabelValues(strconv.FormatBool(valid)).Set(float64(count))
	}

	// Set revocation metrics
	for revoked, count := range revokedCounts {
		e.setupKeysRevoked.WithLabelValues(strconv.FormatBool(revoked)).Set(float64(count))
	}

	// Set ephemeral metrics
	for ephemeral, count := range ephemeralCounts {
		e.setupKeysEphemeral.WithLabelValues(strconv.FormatBool(ephemeral)).Set(float64(count))
	}

	logrus.WithFields(logrus.Fields{
		"total_keys":   totalKeys,
		"valid_keys":   validCounts[true],
		"revoked_keys": revokedCounts[true],
	}).Debug("Updated setup key metrics")
}
