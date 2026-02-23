package exporters

import (
	"context"
	"time"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// RoutesExporter handles routes-specific metrics collection
type RoutesExporter struct {
	client *nbclient.Client

	// Prometheus metrics for routes
	routesTotal          *prometheus.GaugeVec
	routesByStatus       *prometheus.GaugeVec
	routesByNetworkType  *prometheus.GaugeVec
	routeInfo            *prometheus.GaugeVec
	routeMetric          *prometheus.GaugeVec
	routeGroupsCount     *prometheus.GaugeVec
	scrapeErrorsTotal    *prometheus.CounterVec
	scrapeDuration       *prometheus.HistogramVec
}

// NewRoutesExporter creates a new routes exporter
func NewRoutesExporter(client *nbclient.Client) *RoutesExporter {
	return &RoutesExporter{
		client: client,

		routesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_routes",
				Help: "Total number of NetBird routes",
			},
			[]string{},
		),

		routesByStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_routes_by_status",
				Help: "Number of NetBird routes by status (enabled/disabled)",
			},
			[]string{"enabled"},
		),

		routesByNetworkType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_routes_by_network_type",
				Help: "Number of NetBird routes by network type",
			},
			[]string{"network_type"},
		),

		routeInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_route_info",
				Help: "Information about NetBird routes (always 1)",
			},
			[]string{"route_id", "network_id", "network_type", "enabled", "masquerade", "keep_route"},
		),

		routeMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_route_metric_value",
				Help: "Metric value for NetBird route (lower = higher priority)",
			},
			[]string{"route_id", "network_id"},
		),

		routeGroupsCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_route_groups_count",
				Help: "Number of peer groups associated with each NetBird route",
			},
			[]string{"route_id", "network_id"},
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
	e.routesByStatus.Describe(ch)
	e.routesByNetworkType.Describe(ch)
	e.routeInfo.Describe(ch)
	e.routeMetric.Describe(ch)
	e.routeGroupsCount.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *RoutesExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.routesTotal.Reset()
	e.routesByStatus.Reset()
	e.routesByNetworkType.Reset()
	e.routeInfo.Reset()
	e.routeMetric.Reset()
	e.routeGroupsCount.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
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
	e.routesByStatus.Collect(ch)
	e.routesByNetworkType.Collect(ch)
	e.routeInfo.Collect(ch)
	e.routeMetric.Collect(ch)
	e.routeGroupsCount.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on routes data
func (e *RoutesExporter) updateMetrics(routes []api.Route) {
	enabledCount := 0
	disabledCount := 0
	networkTypeCount := make(map[string]int)

	for _, route := range routes {
		if route.Enabled {
			enabledCount++
		} else {
			disabledCount++
		}

		// Count by network type
		networkTypeCount[route.NetworkType]++

		// Convert booleans to strings
		enabledStr := "false"
		if route.Enabled {
			enabledStr = "true"
		}
		masqueradeStr := "false"
		if route.Masquerade {
			masqueradeStr = "true"
		}
		keepRouteStr := "false"
		if route.KeepRoute {
			keepRouteStr = "true"
		}

		// Set route info metric
		infoLabels := []string{route.Id, route.NetworkId, route.NetworkType, enabledStr, masqueradeStr, keepRouteStr}
		e.routeInfo.WithLabelValues(infoLabels...).Set(1)

		// Set route metric value
		metricLabels := []string{route.Id, route.NetworkId}
		e.routeMetric.WithLabelValues(metricLabels...).Set(float64(route.Metric))

		// Set groups count
		e.routeGroupsCount.WithLabelValues(metricLabels...).Set(float64(len(route.Groups)))
	}

	// Set total and distribution metrics
	e.routesTotal.WithLabelValues().Set(float64(len(routes)))
	e.routesByStatus.WithLabelValues("true").Set(float64(enabledCount))
	e.routesByStatus.WithLabelValues("false").Set(float64(disabledCount))

	for networkType, count := range networkTypeCount {
		e.routesByNetworkType.WithLabelValues(networkType).Set(float64(count))
	}

	logrus.WithFields(logrus.Fields{
		"total_routes":     len(routes),
		"enabled":          enabledCount,
		"disabled":         disabledCount,
		"by_network_type":  networkTypeCount,
	}).Debug("Updated routes metrics")
}
