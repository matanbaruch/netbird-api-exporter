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

// PoliciesExporter handles policy-specific metrics collection
type PoliciesExporter struct {
	client *nbclient.Client

	// Prometheus metrics for policies
	policiesTotal       *prometheus.GaugeVec
	policyRulesCount    *prometheus.GaugeVec
	policyRulesEnabled  *prometheus.GaugeVec
	policyRulesByProto  *prometheus.GaugeVec
	policyRulesByAction *prometheus.GaugeVec
	policyInfo          *prometheus.GaugeVec
	scrapeErrorsTotal   *prometheus.CounterVec
	scrapeDuration      *prometheus.HistogramVec
}

// NewPoliciesExporter creates a new policies exporter
func NewPoliciesExporter(client *nbclient.Client) *PoliciesExporter {
	return &PoliciesExporter{
		client: client,

		policiesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policies",
				Help: "Total number of NetBird policies grouped by enabled status",
			},
			[]string{"enabled"},
		),

		policyRulesCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policy_rules_count",
				Help: "Number of rules configured in each NetBird policy",
			},
			[]string{"policy_id", "policy_name"},
		),

		policyRulesEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policy_rules_enabled",
				Help: "Number of NetBird policy rules grouped by enabled status",
			},
			[]string{"enabled"},
		),

		policyRulesByProto: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policy_rules_by_protocol",
				Help: "Number of NetBird policy rules grouped by protocol",
			},
			[]string{"protocol"},
		),

		policyRulesByAction: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policy_rules_by_action",
				Help: "Number of NetBird policy rules grouped by action",
			},
			[]string{"action"},
		),

		policyInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_policy_info",
				Help: "Information about NetBird policies (always 1)",
			},
			[]string{"policy_id", "policy_name", "description"},
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
	e.policyRulesCount.Describe(ch)
	e.policyRulesEnabled.Describe(ch)
	e.policyRulesByProto.Describe(ch)
	e.policyRulesByAction.Describe(ch)
	e.policyInfo.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *PoliciesExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.policiesTotal.Reset()
	e.policyRulesCount.Reset()
	e.policyRulesEnabled.Reset()
	e.policyRulesByProto.Reset()
	e.policyRulesByAction.Reset()
	e.policyInfo.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	e.policyRulesCount.Collect(ch)
	e.policyRulesEnabled.Collect(ch)
	e.policyRulesByProto.Collect(ch)
	e.policyRulesByAction.Collect(ch)
	e.policyInfo.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on policies data
func (e *PoliciesExporter) updateMetrics(policies []api.Policy) {
	enabledCounts := make(map[bool]int)
	ruleEnabledCounts := make(map[bool]int)
	protocolCounts := make(map[string]int)
	actionCounts := make(map[string]int)
	totalRules := 0

	for _, policy := range policies {
		enabledCounts[policy.Enabled]++

		policyID := ""
		if policy.Id != nil {
			policyID = *policy.Id
		}
		description := ""
		if policy.Description != nil {
			description = *policy.Description
		}

		e.policyRulesCount.WithLabelValues(policyID, policy.Name).Set(float64(len(policy.Rules)))
		e.policyInfo.WithLabelValues(policyID, policy.Name, description).Set(1)

		for _, rule := range policy.Rules {
			ruleEnabledCounts[rule.Enabled]++
			protocolCounts[string(rule.Protocol)]++
			actionCounts[string(rule.Action)]++
			totalRules++
		}
	}

	// Set policy totals by enabled status
	for enabled, count := range enabledCounts {
		e.policiesTotal.WithLabelValues(strconv.FormatBool(enabled)).Set(float64(count))
	}

	// Set rule enabled metrics
	for enabled, count := range ruleEnabledCounts {
		e.policyRulesEnabled.WithLabelValues(strconv.FormatBool(enabled)).Set(float64(count))
	}

	// Set rule protocol metrics
	for protocol, count := range protocolCounts {
		e.policyRulesByProto.WithLabelValues(protocol).Set(float64(count))
	}

	// Set rule action metrics
	for action, count := range actionCounts {
		e.policyRulesByAction.WithLabelValues(action).Set(float64(count))
	}

	logrus.WithFields(logrus.Fields{
		"total_policies": len(policies),
		"total_rules":    totalRules,
	}).Debug("Updated policy metrics")
}
