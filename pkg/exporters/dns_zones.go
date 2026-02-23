package exporters

import (
	"context"
	"time"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// DNSZonesExporter handles DNS zones-specific metrics collection
type DNSZonesExporter struct {
	client *nbclient.Client

	// Prometheus metrics for DNS zones
	dnsZonesTotal       *prometheus.GaugeVec
	dnsZoneInfo         *prometheus.GaugeVec
	dnsZoneRecordsCount *prometheus.GaugeVec
	scrapeErrorsTotal   *prometheus.CounterVec
	scrapeDuration      *prometheus.HistogramVec
}

// NewDNSZonesExporter creates a new DNS zones exporter
func NewDNSZonesExporter(client *nbclient.Client) *DNSZonesExporter {
	return &DNSZonesExporter{
		client: client,

		dnsZonesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_zones",
				Help: "Total number of NetBird DNS zones",
			},
			[]string{},
		),

		dnsZoneInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_zone_info",
				Help: "Information about NetBird DNS zones (always 1)",
			},
			[]string{"zone_id", "zone_name", "domain", "enabled"},
		),

		dnsZoneRecordsCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_zone_records_count",
				Help: "Number of DNS records in each NetBird DNS zone",
			},
			[]string{"zone_id", "zone_name"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_dns_zones_scrape_errors_total",
				Help: "Total number of errors encountered while scraping DNS zones",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_dns_zones_scrape_duration_seconds",
				Help: "Time spent scraping DNS zones from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *DNSZonesExporter) Describe(ch chan<- *prometheus.Desc) {
	e.dnsZonesTotal.Describe(ch)
	e.dnsZoneInfo.Describe(ch)
	e.dnsZoneRecordsCount.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *DNSZonesExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.dnsZonesTotal.Reset()
	e.dnsZoneInfo.Reset()
	e.dnsZoneRecordsCount.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	zones, err := e.client.DNSZones.ListZones(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch DNS zones")
		e.scrapeErrorsTotal.WithLabelValues("fetch_dns_zones").Inc()
		return
	}

	e.updateMetrics(ctx, zones)

	// Collect all metrics
	e.dnsZonesTotal.Collect(ch)
	e.dnsZoneInfo.Collect(ch)
	e.dnsZoneRecordsCount.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on DNS zones data
func (e *DNSZonesExporter) updateMetrics(ctx context.Context, zones []api.Zone) {
	for _, zone := range zones {
		// Convert enabled to string
		enabledStr := "false"
		if zone.Enabled {
			enabledStr = "true"
		}

		// Set zone info metric
		infoLabels := []string{zone.Id, zone.Name, zone.Domain, enabledStr}
		e.dnsZoneInfo.WithLabelValues(infoLabels...).Set(1)

		// Fetch records count for this zone
		records, err := e.client.DNSZones.ListRecords(ctx, zone.Id)
		if err != nil {
			logrus.WithError(err).WithField("zone_id", zone.Id).Error("Failed to fetch DNS records for zone")
			e.scrapeErrorsTotal.WithLabelValues("fetch_dns_records").Inc()
			continue
		}

		// Set records count
		countLabels := []string{zone.Id, zone.Name}
		e.dnsZoneRecordsCount.WithLabelValues(countLabels...).Set(float64(len(records)))
	}

	// Set total metric
	e.dnsZonesTotal.WithLabelValues().Set(float64(len(zones)))

	logrus.WithFields(logrus.Fields{
		"total_dns_zones": len(zones),
	}).Debug("Updated DNS zones metrics")
}
