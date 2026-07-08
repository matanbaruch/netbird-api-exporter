package exporters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewPoliciesExporter(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewPoliciesExporter(client)

	if exporter == nil {
		t.Fatal("Expected exporter to be non-nil")
	}

	if exporter.client != client {
		t.Error("Expected client to be set correctly")
	}

	if exporter.policiesTotal == nil {
		t.Error("Expected policiesTotal metric to be non-nil")
	}
	if exporter.policyRulesCount == nil {
		t.Error("Expected policyRulesCount metric to be non-nil")
	}
	if exporter.policyRulesByProto == nil {
		t.Error("Expected policyRulesByProto metric to be non-nil")
	}
	if exporter.policyRulesByAction == nil {
		t.Error("Expected policyRulesByAction metric to be non-nil")
	}
}

func TestPoliciesExporter_Describe(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewPoliciesExporter(client)

	ch := make(chan *prometheus.Desc, 20)
	go func() {
		exporter.Describe(ch)
		close(ch)
	}()

	count := 0
	for desc := range ch {
		if desc == nil {
			t.Error("Expected metric description to be non-nil")
		}
		count++
	}

	if count == 0 {
		t.Error("Expected at least one metric description")
	}
}

func testPolicies() []api.Policy {
	id1 := "policy1"
	id2 := "policy2"
	desc1 := "Allow all"
	ruleID1 := "rule1"
	ruleID2 := "rule2"
	return []api.Policy{
		{
			Id:          &id1,
			Name:        "allow-all",
			Description: &desc1,
			Enabled:     true,
			Rules: []api.PolicyRule{
				{
					Id:            &ruleID1,
					Name:          "rule1",
					Enabled:       true,
					Action:        api.PolicyRuleActionAccept,
					Protocol:      api.PolicyRuleProtocolAll,
					Bidirectional: true,
				},
			},
		},
		{
			Id:      &id2,
			Name:    "deny-tcp",
			Enabled: false,
			Rules: []api.PolicyRule{
				{
					Id:            &ruleID2,
					Name:          "rule2",
					Enabled:       false,
					Action:        api.PolicyRuleActionDrop,
					Protocol:      api.PolicyRuleProtocolTcp,
					Bidirectional: false,
				},
			},
		},
	}
}

func TestPoliciesExporter_Collect_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/policies" {
			http.NotFound(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Token ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(testPolicies()); err != nil {
			t.Errorf("Failed to encode policies: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewPoliciesExporter(client)

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	protoFound := false
	for _, family := range families {
		if family.GetName() == "netbird_policy_rules_by_protocol" {
			protoFound = true
			for _, m := range family.GetMetric() {
				for _, l := range m.GetLabel() {
					if l.GetName() == "protocol" && l.GetValue() == "tcp" {
						if m.GetGauge().GetValue() != 1 {
							t.Errorf("Expected 1 tcp rule, got %f", m.GetGauge().GetValue())
						}
					}
				}
			}
		}
	}

	if !protoFound {
		t.Error("Expected to find netbird_policy_rules_by_protocol metric")
	}
}

func TestPoliciesExporter_Collect_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewPoliciesExporter(client)

	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	for range ch {
		// Drain channel; must not panic or hang
	}
}

func TestPoliciesExporter_Collect_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewPoliciesExporter(client)

	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("Expected at least one metric (even if zero)")
	}
}

func TestPoliciesExporter_Collect_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`invalid json`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewPoliciesExporter(client)

	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	for range ch {
		// Drain channel
	}
}

func TestPoliciesExporter_UpdateMetrics(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewPoliciesExporter(client)

	exporter.updateMetrics(testPolicies())

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		exporter.policiesTotal,
		exporter.policyRulesCount,
		exporter.policyRulesEnabled,
		exporter.policyRulesByProto,
		exporter.policyRulesByAction,
	)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() == "netbird_policies" {
			for _, m := range family.GetMetric() {
				for _, l := range m.GetLabel() {
					if l.GetName() == "enabled" && l.GetValue() == "true" {
						if m.GetGauge().GetValue() != 1 {
							t.Errorf("Expected 1 enabled policy, got %f", m.GetGauge().GetValue())
						}
					}
				}
			}
		}
	}
}

func TestPoliciesExporter_MetricsReset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewPoliciesExporter(client)

	// Set a stale series that should be cleared on the next collection
	exporter.policyRulesCount.WithLabelValues("policy1", "stale-policy").Set(99)

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter.policyRulesCount)

	// Collect resets metrics before fetching; the empty response clears stale series
	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()
	for range ch {
		// Drain channel
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() == "netbird_policy_rules_count" {
			if len(family.GetMetric()) != 0 {
				t.Errorf("Expected policy rules count to be reset, got %d metrics", len(family.GetMetric()))
			}
		}
	}
}
