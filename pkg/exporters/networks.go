package exporters

import (
	"context"
	"time"

	nbclient "github.com/netbirdio/netbird/management/client/rest"
	"github.com/netbirdio/netbird/management/server/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// NetworksExporter handles networks-specific metrics collection
type NetworksExporter struct {
	client *nbclient.Client

	// Prometheus metrics for networks
	networksTotal            *prometheus.GaugeVec
	networkRoutersCount      *prometheus.GaugeVec
	networkResourcesCount    *prometheus.GaugeVec
	networkPoliciesCount     *prometheus.GaugeVec
	networkRoutingPeersCount *prometheus.GaugeVec
	networkInfo              *prometheus.GaugeVec
	scrapeErrorsTotal        *prometheus.CounterVec
	scrapeDuration           *prometheus.HistogramVec
}

// NewNetworksExporter creates a new networks exporter
func NewNetworksExporter(client *nbclient.Client) *NetworksExporter {
	return &NetworksExporter{
		client: client,

		networksTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_networks",
				Help: "Total number of NetBird networks",
			},
			[]string{},
		),

		networkRoutersCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_network_routers_count",
				Help: "Number of routers in each NetBird network",
			},
			[]string{"network_id", "network_name"},
		),

		networkResourcesCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_network_resources_count",
				Help: "Number of resources in each NetBird network",
			},
			[]string{"network_id", "network_name"},
		),

		networkPoliciesCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_network_policies_count",
				Help: "Number of policies in each NetBird network",
			},
			[]string{"network_id", "network_name"},
		),

		networkRoutingPeersCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_network_routing_peers_count",
				Help: "Number of routing peers in each NetBird network",
			},
			[]string{"network_id", "network_name"},
		),

		networkInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_network_info",
				Help: "Information about NetBird networks (always 1)",
			},
			[]string{"network_id", "network_name", "description"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_networks_scrape_errors_total",
				Help: "Total number of errors encountered while scraping networks",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_networks_scrape_duration_seconds",
				Help: "Time spent scraping networks from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *NetworksExporter) Describe(ch chan<- *prometheus.Desc) {
	e.networksTotal.Describe(ch)
	e.networkRoutersCount.Describe(ch)
	e.networkResourcesCount.Describe(ch)
	e.networkPoliciesCount.Describe(ch)
	e.networkRoutingPeersCount.Describe(ch)
	e.networkInfo.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *NetworksExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.networksTotal.Reset()
	e.networkRoutersCount.Reset()
	e.networkResourcesCount.Reset()
	e.networkPoliciesCount.Reset()
	e.networkRoutingPeersCount.Reset()
	e.networkInfo.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	networks, err := e.client.Networks.List(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch networks")
		e.scrapeErrorsTotal.WithLabelValues("fetch_networks").Inc()
		return
	}

	e.updateMetrics(networks)

	// Collect all metrics
	e.networksTotal.Collect(ch)
	e.networkRoutersCount.Collect(ch)
	e.networkResourcesCount.Collect(ch)
	e.networkPoliciesCount.Collect(ch)
	e.networkRoutingPeersCount.Collect(ch)
	e.networkInfo.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on networks data
func (e *NetworksExporter) updateMetrics(networks []api.Network) {
	totalNetworks := len(networks)
	totalRouters := 0
	totalResources := 0
	totalPolicies := 0
	totalRoutingPeers := 0

	for _, network := range networks {
		networkLabels := []string{network.Id, network.Name}
		description := ""
		if network.Description != nil {
			description = *network.Description
		}
		infoLabels := []string{network.Id, network.Name, description}

		routersCount := len(network.Routers)
		resourcesCount := len(network.Resources)
		policiesCount := len(network.Policies)

		// Set basic network metrics
		e.networkRoutersCount.WithLabelValues(networkLabels...).Set(float64(routersCount))
		e.networkResourcesCount.WithLabelValues(networkLabels...).Set(float64(resourcesCount))
		e.networkPoliciesCount.WithLabelValues(networkLabels...).Set(float64(policiesCount))
		e.networkRoutingPeersCount.WithLabelValues(networkLabels...).Set(float64(network.RoutingPeersCount))
		e.networkInfo.WithLabelValues(infoLabels...).Set(1)

		// Add to totals
		totalRouters += routersCount
		totalResources += resourcesCount
		totalPolicies += policiesCount
		totalRoutingPeers += network.RoutingPeersCount
	}

	e.networksTotal.WithLabelValues().Set(float64(totalNetworks))

	logrus.WithFields(logrus.Fields{
		"total_networks":      totalNetworks,
		"total_routers":       totalRouters,
		"total_resources":     totalResources,
		"total_policies":      totalPolicies,
		"total_routing_peers": totalRoutingPeers,
	}).Debug("Updated network metrics")
}
