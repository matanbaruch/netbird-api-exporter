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

// EventsExporter handles events-specific metrics collection
// NOTE: This exporter is opt-in due to high cardinality
type EventsExporter struct {
	client  *nbclient.Client
	enabled bool

	// Prometheus metrics for events
	eventsTotal        *prometheus.GaugeVec
	eventsByActivity   *prometheus.GaugeVec
	eventInfo          *prometheus.GaugeVec
	eventTimestamp     *prometheus.GaugeVec
	scrapeErrorsTotal  *prometheus.CounterVec
	scrapeDuration     *prometheus.HistogramVec
}

// NewEventsExporter creates a new events exporter
// This exporter is disabled by default and must be explicitly enabled via ENABLE_EVENTS_EXPORTER=true
func NewEventsExporter(client *nbclient.Client) *EventsExporter {
	enabled := os.Getenv("ENABLE_EVENTS_EXPORTER") == "true"

	if enabled {
		logrus.Warn("Events exporter is enabled (HIGH cardinality metrics - use with caution)")
	} else {
		logrus.Info("Events exporter is disabled (set ENABLE_EVENTS_EXPORTER=true to enable)")
	}

	return &EventsExporter{
		client:  client,
		enabled: enabled,

		eventsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_events",
				Help: "Total number of NetBird events retrieved",
			},
			[]string{},
		),

		eventsByActivity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_events_by_activity",
				Help: "Number of NetBird events by activity type",
			},
			[]string{"activity_code"},
		),

		eventInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_event_info",
				Help: "Information about NetBird events (always 1)",
			},
			[]string{"event_id", "activity", "activity_code", "initiator_id", "initiator_email", "target_id"},
		),

		eventTimestamp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_event_timestamp",
				Help: "Unix timestamp when the event occurred",
			},
			[]string{"event_id", "activity_code"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_events_scrape_errors_total",
				Help: "Total number of errors encountered while scraping events",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_events_scrape_duration_seconds",
				Help: "Time spent scraping events from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *EventsExporter) Describe(ch chan<- *prometheus.Desc) {
	if !e.enabled {
		return
	}

	e.eventsTotal.Describe(ch)
	e.eventsByActivity.Describe(ch)
	e.eventInfo.Describe(ch)
	e.eventTimestamp.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *EventsExporter) Collect(ch chan<- prometheus.Metric) {
	if !e.enabled {
		return
	}

	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.eventsTotal.Reset()
	e.eventsByActivity.Reset()
	e.eventInfo.Reset()
	e.eventTimestamp.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	events, err := e.client.Events.List(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch events")
		e.scrapeErrorsTotal.WithLabelValues("fetch_events").Inc()
		return
	}

	e.updateMetrics(events)

	// Collect all metrics
	e.eventsTotal.Collect(ch)
	e.eventsByActivity.Collect(ch)
	e.eventInfo.Collect(ch)
	e.eventTimestamp.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on events data
func (e *EventsExporter) updateMetrics(events []api.Event) {
	activityCount := make(map[string]int)

	for _, event := range events {
		// Count by activity code
		activityCodeStr := string(event.ActivityCode)
		activityCount[activityCodeStr]++

		// Set event info metric
		infoLabels := []string{
			event.Id,
			event.Activity,
			activityCodeStr,
			event.InitiatorId,
			event.InitiatorEmail,
			event.TargetId,
		}
		e.eventInfo.WithLabelValues(infoLabels...).Set(1)

		// Set event timestamp
		timestampLabels := []string{event.Id, activityCodeStr}
		e.eventTimestamp.WithLabelValues(timestampLabels...).Set(float64(event.Timestamp.Unix()))
	}

	// Set total events
	e.eventsTotal.WithLabelValues().Set(float64(len(events)))

	// Set events by activity
	for activityCode, count := range activityCount {
		e.eventsByActivity.WithLabelValues(activityCode).Set(float64(count))
	}

	logrus.WithFields(logrus.Fields{
		"total_events":   len(events),
		"activity_types": len(activityCount),
	}).Debug("Updated events metrics")
}
