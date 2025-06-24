package exporters

import (
	"context"
	"strconv"
	"time"

	nbclient "github.com/netbirdio/netbird/management/client/rest"
	"github.com/netbirdio/netbird/management/server/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// DNSExporter handles DNS-specific metrics collection
type DNSExporter struct {
	client *nbclient.Client

	// Prometheus metrics
	nameserverGroupsTotal   *prometheus.GaugeVec
	nameserverGroupsEnabled *prometheus.GaugeVec
	nameserverGroupsPrimary *prometheus.GaugeVec
	nameserverGroupDomains  *prometheus.GaugeVec
	nameserversTotal        *prometheus.GaugeVec
	nameserversByType       *prometheus.GaugeVec
	nameserversByPort       *prometheus.GaugeVec
	dnsManagementDisabled   *prometheus.GaugeVec
}

// NewDNSExporter creates a new DNS exporter
func NewDNSExporter(client *nbclient.Client) *DNSExporter {
	return &DNSExporter{
		client: client,

		nameserverGroupsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_nameserver_groups",
				Help: "Total number of NetBird nameserver groups",
			},
			[]string{},
		),

		nameserverGroupsEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_nameserver_groups_enabled",
				Help: "Number of enabled NetBird nameserver groups",
			},
			[]string{"enabled"},
		),

		nameserverGroupsPrimary: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_nameserver_groups_primary",
				Help: "Number of primary NetBird nameserver groups",
			},
			[]string{"primary"},
		),

		nameserverGroupDomains: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_nameserver_group_domains_count",
				Help: "Number of domains configured in each nameserver group",
			},
			[]string{"group_id", "group_name"},
		),

		nameserversTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_nameservers",
				Help: "Total number of nameservers across all groups",
			},
			[]string{"group_id", "group_name"},
		),

		nameserversByType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_nameservers_by_type",
				Help: "Number of nameservers by type (UDP/TCP)",
			},
			[]string{"ns_type"},
		),

		nameserversByPort: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_nameservers_by_port",
				Help: "Number of nameservers by port",
			},
			[]string{"port"},
		),

		dnsManagementDisabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_dns_management_disabled_groups_count",
				Help: "Number of groups with DNS management disabled",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *DNSExporter) Describe(ch chan<- *prometheus.Desc) {
	e.nameserverGroupsTotal.Describe(ch)
	e.nameserverGroupsEnabled.Describe(ch)
	e.nameserverGroupsPrimary.Describe(ch)
	e.nameserverGroupDomains.Describe(ch)
	e.nameserversTotal.Describe(ch)
	e.nameserversByType.Describe(ch)
	e.nameserversByPort.Describe(ch)
	e.dnsManagementDisabled.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *DNSExporter) Collect(ch chan<- prometheus.Metric) {
	// Reset metrics before collecting new values
	e.nameserverGroupsTotal.Reset()
	e.nameserverGroupsEnabled.Reset()
	e.nameserverGroupsPrimary.Reset()
	e.nameserverGroupDomains.Reset()
	e.nameserversTotal.Reset()
	e.nameserversByType.Reset()
	e.nameserversByPort.Reset()
	e.dnsManagementDisabled.Reset()

	ctx, cancelNS := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelNS()
	// Fetch nameserver groups
	nameserverGroups, err := e.client.DNS.ListNameserverGroups(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch nameserver groups")
	} else {
		e.updateNameserverMetrics(nameserverGroups)
	}

	ctx, cancelSettings := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelSettings()

	// Fetch DNS settings
	dnsSettings, err := e.client.DNS.GetSettings(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch DNS settings")
	} else {
		e.updateDNSSettingsMetrics(dnsSettings)
	}

	// Collect all metrics
	e.nameserverGroupsTotal.Collect(ch)
	e.nameserverGroupsEnabled.Collect(ch)
	e.nameserverGroupsPrimary.Collect(ch)
	e.nameserverGroupDomains.Collect(ch)
	e.nameserversTotal.Collect(ch)
	e.nameserversByType.Collect(ch)
	e.nameserversByPort.Collect(ch)
	e.dnsManagementDisabled.Collect(ch)
}

// updateNameserverMetrics updates Prometheus metrics based on nameserver group data
func (e *DNSExporter) updateNameserverMetrics(nameserverGroups []api.NameserverGroup) {
	// Count totals
	totalGroups := len(nameserverGroups)
	e.nameserverGroupsTotal.WithLabelValues().Set(float64(totalGroups))

	// Count by status
	enabledCounts := make(map[bool]int)
	primaryCounts := make(map[bool]int)
	typeCounter := make(map[string]int)
	portCounter := make(map[string]int)

	for _, group := range nameserverGroups {
		// Count enabled/disabled
		enabledCounts[group.Enabled]++

		// Count primary/secondary
		primaryCounts[group.Primary]++

		// Count domains per group
		e.nameserverGroupDomains.WithLabelValues(group.Id, group.Name).Set(float64(len(group.Domains)))

		// Count nameservers per group
		e.nameserversTotal.WithLabelValues(group.Id, group.Name).Set(float64(len(group.Nameservers)))

		// Count nameserver types and ports
		for _, ns := range group.Nameservers {
			typeCounter[string(ns.NsType)]++
			portCounter[strconv.Itoa(ns.Port)]++
		}
	}

	// Set enabled/disabled metrics
	for enabled, count := range enabledCounts {
		e.nameserverGroupsEnabled.WithLabelValues(strconv.FormatBool(enabled)).Set(float64(count))
	}

	// Set primary/secondary metrics
	for primary, count := range primaryCounts {
		e.nameserverGroupsPrimary.WithLabelValues(strconv.FormatBool(primary)).Set(float64(count))
	}

	// Set nameserver type metrics
	for nsType, count := range typeCounter {
		e.nameserversByType.WithLabelValues(nsType).Set(float64(count))
	}

	// Set nameserver port metrics
	for port, count := range portCounter {
		e.nameserversByPort.WithLabelValues(port).Set(float64(count))
	}

	logrus.WithFields(logrus.Fields{
		"total_groups":    totalGroups,
		"enabled_groups":  enabledCounts[true],
		"disabled_groups": enabledCounts[false],
		"primary_groups":  primaryCounts[true],
	}).Debug("Updated nameserver metrics")
}

// updateDNSSettingsMetrics updates Prometheus metrics based on DNS settings data
func (e *DNSExporter) updateDNSSettingsMetrics(dnsSettings *api.DNSSettings) {
	// Count disabled management groups
	disabledCount := len(dnsSettings.DisabledManagementGroups)
	e.dnsManagementDisabled.WithLabelValues().Set(float64(disabledCount))

	logrus.WithField("disabled_management_groups", disabledCount).Debug("Updated DNS settings metrics")
}
