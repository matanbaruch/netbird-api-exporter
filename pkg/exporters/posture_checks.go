package exporters

import (
	"context"
	"time"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// PostureChecksExporter handles posture checks-specific metrics collection
type PostureChecksExporter struct {
	client *nbclient.Client

	// Prometheus metrics for posture checks
	postureChecksTotal  *prometheus.GaugeVec
	postureCheckInfo    *prometheus.GaugeVec
	scrapeErrorsTotal   *prometheus.CounterVec
	scrapeDuration      *prometheus.HistogramVec
}

// NewPostureChecksExporter creates a new posture checks exporter
func NewPostureChecksExporter(client *nbclient.Client) *PostureChecksExporter {
	return &PostureChecksExporter{
		client: client,

		postureChecksTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_posture_checks",
				Help: "Total number of NetBird posture checks",
			},
			[]string{},
		),

		postureCheckInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_posture_check_info",
				Help: "Information about NetBird posture checks (always 1)",
			},
			[]string{"check_id", "check_name"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_posture_checks_scrape_errors_total",
				Help: "Total number of errors encountered while scraping posture checks",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_posture_checks_scrape_duration_seconds",
				Help: "Time spent scraping posture checks from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *PostureChecksExporter) Describe(ch chan<- *prometheus.Desc) {
	e.postureChecksTotal.Describe(ch)
	e.postureCheckInfo.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *PostureChecksExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.postureChecksTotal.Reset()
	e.postureCheckInfo.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	postureChecks, err := e.client.PostureChecks.List(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch posture checks")
		e.scrapeErrorsTotal.WithLabelValues("fetch_posture_checks").Inc()
		return
	}

	e.updateMetrics(postureChecks)

	// Collect all metrics
	e.postureChecksTotal.Collect(ch)
	e.postureCheckInfo.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on posture checks data
func (e *PostureChecksExporter) updateMetrics(postureChecks []api.PostureCheck) {
	for _, check := range postureChecks {
		// Set posture check info metric
		infoLabels := []string{check.Id, check.Name}
		e.postureCheckInfo.WithLabelValues(infoLabels...).Set(1)
	}

	// Set total metric
	e.postureChecksTotal.WithLabelValues().Set(float64(len(postureChecks)))

	logrus.WithFields(logrus.Fields{
		"total_posture_checks": len(postureChecks),
	}).Debug("Updated posture checks metrics")
}
