package exporters

import (
	"context"
	"time"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// PoliciesExporter handles policies-specific metrics collection
type PoliciesExporter struct {
	client *nbclient.Client

	// Prometheus metrics for policies
	policiesTotal         *prometheus.GaugeVec
	policiesByStatus      *prometheus.GaugeVec
	policyInfo            *prometheus.GaugeVec
	policyRulesCount      *prometheus.GaugeVec
	policyPostureChecks   *prometheus.GaugeVec
	scrapeErrorsTotal     *prometheus.CounterVec
	scrapeDuration        *prometheus.HistogramVec
}

// NewPoliciesExporter creates a new policies exporter
func NewPoliciesExporter(client *nbclient.Client) *PoliciesExporter {
	return &PoliciesExporter{
		client: client,

		policiesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policies",
				Help: "Total number of NetBird policies",
			},
			[]string{},
		),

		policiesByStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policies_by_status",
				Help: "Number of NetBird policies by status (enabled/disabled)",
			},
			[]string{"enabled"},
		),

		policyInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policy_info",
				Help: "Information about NetBird policies (always 1)",
			},
			[]string{"policy_id", "policy_name", "enabled"},
		),

		policyRulesCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policy_rules_count",
				Help: "Number of rules in each NetBird policy",
			},
			[]string{"policy_id", "policy_name"},
		),

		policyPostureChecks: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policy_posture_checks_count",
				Help: "Number of source posture checks in each NetBird policy",
			},
			[]string{"policy_id", "policy_name"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_policies_scrape_errors_total",
				Help: "Total number of errors encountered while scraping policies",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_policies_scrape_duration_seconds",
				Help: "Time spent scraping policies from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *PoliciesExporter) Describe(ch chan<- *prometheus.Desc) {
	e.policiesTotal.Describe(ch)
	e.policiesByStatus.Describe(ch)
	e.policyInfo.Describe(ch)
	e.policyRulesCount.Describe(ch)
	e.policyPostureChecks.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *PoliciesExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.policiesTotal.Reset()
	e.policiesByStatus.Reset()
	e.policyInfo.Reset()
	e.policyRulesCount.Reset()
	e.policyPostureChecks.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	policies, err := e.client.Policies.List(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch policies")
		e.scrapeErrorsTotal.WithLabelValues("fetch_policies").Inc()
		return
	}

	e.updateMetrics(policies)

	// Collect all metrics
	e.policiesTotal.Collect(ch)
	e.policiesByStatus.Collect(ch)
	e.policyInfo.Collect(ch)
	e.policyRulesCount.Collect(ch)
	e.policyPostureChecks.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on policies data
func (e *PoliciesExporter) updateMetrics(policies []api.Policy) {
	enabledCount := 0
	disabledCount := 0

	for _, policy := range policies {
		if policy.Enabled {
			enabledCount++
		} else {
			disabledCount++
		}

		// Convert enabled to string
		enabledStr := "false"
		if policy.Enabled {
			enabledStr = "true"
		}

		// Get policy ID and name
		policyID := ""
		if policy.Id != nil {
			policyID = *policy.Id
		}

		// Set policy info metric
		infoLabels := []string{policyID, policy.Name, enabledStr}
		e.policyInfo.WithLabelValues(infoLabels...).Set(1)

		// Set rules count
		countLabels := []string{policyID, policy.Name}
		e.policyRulesCount.WithLabelValues(countLabels...).Set(float64(len(policy.Rules)))

		// Set posture checks count
		e.policyPostureChecks.WithLabelValues(countLabels...).Set(float64(len(policy.SourcePostureChecks)))
	}

	// Set total and distribution metrics
	e.policiesTotal.WithLabelValues().Set(float64(len(policies)))
	e.policiesByStatus.WithLabelValues("true").Set(float64(enabledCount))
	e.policiesByStatus.WithLabelValues("false").Set(float64(disabledCount))

	logrus.WithFields(logrus.Fields{
		"total_policies": len(policies),
		"enabled":        enabledCount,
		"disabled":       disabledCount,
	}).Debug("Updated policies metrics")
}
