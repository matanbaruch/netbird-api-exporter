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

// RoutesExporter handles route-specific metrics collection
type RoutesExporter struct {
	client *nbclient.Client

	// Prometheus metrics for routes
	routesTotal         *prometheus.GaugeVec
	routesByNetworkType *prometheus.GaugeVec
	routesMasquerade    *prometheus.GaugeVec
	routeInfo           *prometheus.GaugeVec
	scrapeErrorsTotal   *prometheus.CounterVec
	scrapeDuration      *prometheus.HistogramVec
}

// NewRoutesExporter creates a new routes exporter
func NewRoutesExporter(client *nbclient.Client) *RoutesExporter {
	return &RoutesExporter{
		client: client,

		routesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_routes",
				Help: "Total number of NetBird routes grouped by enabled status",
			},
			[]string{"enabled"},
		),

		routesByNetworkType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_routes_by_network_type",
				Help: "Number of NetBird routes grouped by network type",
			},
			[]string{"network_type"},
		),

		routesMasquerade: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_routes_masquerade",
				Help: "Number of NetBird routes grouped by masquerade status",
			},
			[]string{"masquerade"},
		),

		routeInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_route_info",
				Help: "Information about NetBird routes (always 1)",
			},
			[]string{"route_id", "network_id", "network_type", "description"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_routes_scrape_errors_total",
				Help: "Total number of errors encountered while scraping routes",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_routes_scrape_duration_seconds",
				Help: "Time spent scraping routes from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *RoutesExporter) Describe(ch chan<- *prometheus.Desc) {
	e.routesTotal.Describe(ch)
	e.routesByNetworkType.Describe(ch)
	e.routesMasquerade.Describe(ch)
	e.routeInfo.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *RoutesExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.routesTotal.Reset()
	e.routesByNetworkType.Reset()
	e.routesMasquerade.Reset()
	e.routeInfo.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	routes, err := e.client.Routes.List(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch routes")
		e.scrapeErrorsTotal.WithLabelValues("fetch_routes").Inc()
		return
	}

	e.updateMetrics(routes)

	// Collect all metrics
	e.routesTotal.Collect(ch)
	e.routesByNetworkType.Collect(ch)
	e.routesMasquerade.Collect(ch)
	e.routeInfo.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on routes data
func (e *RoutesExporter) updateMetrics(routes []api.Route) {
	enabledCounts := make(map[bool]int)
	networkTypeCounts := make(map[string]int)
	masqueradeCounts := make(map[bool]int)

	for _, route := range routes {
		enabledCounts[route.Enabled]++
		networkTypeCounts[route.NetworkType]++
		masqueradeCounts[route.Masquerade]++

		e.routeInfo.WithLabelValues(route.Id, route.NetworkId, route.NetworkType, route.Description).Set(1)
	}

	// Set route totals by enabled status
	for enabled, count := range enabledCounts {
		e.routesTotal.WithLabelValues(strconv.FormatBool(enabled)).Set(float64(count))
	}

	// Set network type metrics
	for networkType, count := range networkTypeCounts {
		e.routesByNetworkType.WithLabelValues(networkType).Set(float64(count))
	}

	// Set masquerade metrics
	for masquerade, count := range masqueradeCounts {
		e.routesMasquerade.WithLabelValues(strconv.FormatBool(masquerade)).Set(float64(count))
	}

	logrus.WithFields(logrus.Fields{
		"total_routes":    len(routes),
		"enabled_routes":  enabledCounts[true],
		"disabled_routes": enabledCounts[false],
	}).Debug("Updated route metrics")
}
