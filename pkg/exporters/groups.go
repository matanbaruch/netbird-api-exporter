package exporters

import (
	"context"
	"time"

	nbclient "github.com/netbirdio/netbird/management/client/rest"
	"github.com/netbirdio/netbird/management/server/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// GroupsExporter handles groups-specific metrics collection
type GroupsExporter struct {
	client *nbclient.Client

	// Prometheus metrics for groups
	groupsTotal          *prometheus.GaugeVec
	groupPeersCount      *prometheus.GaugeVec
	groupResourcesCount  *prometheus.GaugeVec
	groupInfo            *prometheus.GaugeVec
	groupResourcesByType *prometheus.GaugeVec
	scrapeErrorsTotal    *prometheus.CounterVec
	scrapeDuration       *prometheus.HistogramVec
}

// NewGroupsExporter creates a new groups exporter
func NewGroupsExporter(client *nbclient.Client) *GroupsExporter {
	return &GroupsExporter{
		client: client,

		groupsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_groups",
				Help: "Total number of NetBird groups",
			},
			[]string{},
		),

		groupPeersCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_group_peers_count",
				Help: "Number of peers in each NetBird group",
			},
			[]string{"group_id", "group_name", "issued"},
		),

		groupResourcesCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_group_resources_count",
				Help: "Number of resources in each NetBird group",
			},
			[]string{"group_id", "group_name", "issued"},
		),

		groupInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_group_info",
				Help: "Information about NetBird groups (always 1)",
			},
			[]string{"group_id", "group_name", "issued"},
		),

		groupResourcesByType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_group_resources_by_type",
				Help: "Number of resources in each NetBird group by resource type",
			},
			[]string{"group_id", "group_name", "resource_type"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_groups_scrape_errors_total",
				Help: "Total number of errors encountered while scraping groups",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_groups_scrape_duration_seconds",
				Help: "Time spent scraping groups from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *GroupsExporter) Describe(ch chan<- *prometheus.Desc) {
	e.groupsTotal.Describe(ch)
	e.groupPeersCount.Describe(ch)
	e.groupResourcesCount.Describe(ch)
	e.groupInfo.Describe(ch)
	e.groupResourcesByType.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *GroupsExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.groupsTotal.Reset()
	e.groupPeersCount.Reset()
	e.groupResourcesCount.Reset()
	e.groupInfo.Reset()
	e.groupResourcesByType.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	groups, err := e.client.Groups.List(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch groups")
		e.scrapeErrorsTotal.WithLabelValues("fetch_groups").Inc()
		return
	}

	e.updateMetrics(groups)

	// Collect all metrics
	e.groupsTotal.Collect(ch)
	e.groupPeersCount.Collect(ch)
	e.groupResourcesCount.Collect(ch)
	e.groupInfo.Collect(ch)
	e.groupResourcesByType.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on groups data
func (e *GroupsExporter) updateMetrics(groups []api.Group) {
	totalGroups := len(groups)

	// Count resources by type across all groups
	resourceTypeCount := make(map[string]map[string]int) // group_id -> resource_type -> count
	totalPeers := 0
	totalResources := 0
	resourceTypeTotals := make(map[string]int)

	for _, group := range groups {
		issued := ""
		if group.Issued != nil {
			issued = string(*group.Issued)
		}
		groupLabels := []string{group.Id, group.Name, issued}

		// Set basic group metrics
		e.groupPeersCount.WithLabelValues(groupLabels...).Set(float64(group.PeersCount))
		e.groupResourcesCount.WithLabelValues(groupLabels...).Set(float64(group.ResourcesCount))
		e.groupInfo.WithLabelValues(groupLabels...).Set(1)

		// Add to totals
		totalPeers += group.PeersCount
		totalResources += group.ResourcesCount

		// Count resources by type for this group
		if resourceTypeCount[group.Id] == nil {
			resourceTypeCount[group.Id] = make(map[string]int)
		}

		for _, resource := range group.Resources {
			resourceTypeCount[group.Id][string(resource.Type)]++
			resourceTypeTotals[string(resource.Type)]++
		}

		// Set resource type metrics
		for resourceType, count := range resourceTypeCount[group.Id] {
			e.groupResourcesByType.WithLabelValues(group.Id, group.Name, resourceType).Set(float64(count))
		}
	}

	e.groupsTotal.WithLabelValues().Set(float64(totalGroups))

	logrus.WithFields(logrus.Fields{
		"total_groups":          totalGroups,
		"total_peers_in_groups": totalPeers,
		"total_resources":       totalResources,
		"resource_types":        len(resourceTypeTotals),
		"resource_type_counts":  resourceTypeTotals,
	}).Debug("Updated group metrics")
}
