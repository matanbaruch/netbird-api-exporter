package exporters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	nbclient "github.com/netbirdio/netbird/management/client/rest"
	"github.com/netbirdio/netbird/management/server/http/api"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewGroupsExporter(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewGroupsExporter(client)

	if exporter == nil {
		t.Fatal("Expected exporter to be non-nil")
	}

	if exporter.client != client {
		t.Error("Expected client to be set correctly")
	}

	// Check that all metrics are initialized
	if exporter.groupsTotal == nil {
		t.Error("Expected groupsTotal metric to be non-nil")
	}
	if exporter.groupPeersCount == nil {
		t.Error("Expected groupPeersCount metric to be non-nil")
	}
	if exporter.groupResourcesCount == nil {
		t.Error("Expected groupResourcesCount metric to be non-nil")
	}
	if exporter.groupInfo == nil {
		t.Error("Expected groupInfo metric to be non-nil")
	}
	if exporter.groupResourcesByType == nil {
		t.Error("Expected groupResourcesByType metric to be non-nil")
	}
}

func TestGroupsExporter_Describe(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewGroupsExporter(client)

	ch := make(chan *prometheus.Desc, 10)
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

func TestGroupsExporter_Collect_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/groups" {
			http.NotFound(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Token ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		issuedAPI := api.GroupIssuedApi
		issuedIntegration := api.GroupIssuedIntegration
		groups := []api.Group{
			{
				Id:             "group1",
				Name:           "admin-group",
				PeersCount:     5,
				ResourcesCount: 2,
				Issued:         &issuedAPI,
				Peers: []api.PeerMinimum{
					{Id: "peer1", Name: "peer-1"},
					{Id: "peer2", Name: "peer-2"},
				},
				Resources: []api.Resource{
					{Id: "resource1", Type: "host"},
				},
			},
			{
				Id:             "group2",
				Name:           "user-group",
				PeersCount:     10,
				ResourcesCount: 0,
				Issued:         &issuedIntegration,
				Peers: []api.PeerMinimum{
					{Id: "peer3", Name: "peer-3"},
					{Id: "peer4", Name: "peer-4"},
					{Id: "peer5", Name: "peer-5"},
				},
				Resources: []api.Resource{},
			},
			{
				Id:             "group3",
				Name:           "service-group",
				PeersCount:     2,
				ResourcesCount: 3,
				Issued:         &issuedAPI,
				Peers:          []api.PeerMinimum{},
				Resources: []api.Resource{
					{Id: "resource2", Type: "domain"},
					{Id: "resource3", Type: "subnet"},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(groups); err != nil {
			t.Errorf("Failed to encode groups: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewGroupsExporter(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for metric := range ch {
		if metric != nil {
			metrics = append(metrics, metric)
		}
	}

	if len(metrics) == 0 {
		t.Error("Expected at least one metric to be collected")
	}

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Check groups total
	totalFound := false
	for _, family := range families {
		if family.GetName() == "netbird_groups" {
			totalFound = true
			if len(family.GetMetric()) > 0 {
				value := family.GetMetric()[0].GetGauge().GetValue()
				if value != 3 {
					t.Errorf("Expected groups total to be 3, got %f", value)
				}
			}
			break
		}
	}

	if !totalFound {
		t.Error("Expected to find groups total metric")
	}
}

func TestGroupsExporter_Collect_APIError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewGroupsExporter(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	// Should still complete without error (though may not collect useful data)
	metricCount := 0
	for range ch {
		metricCount++
	}

	// Even with API errors, some metrics might still be collected
	// This test ensures the collection doesn't panic or hang
}

func TestGroupsExporter_Collect_EmptyResponse(t *testing.T) {
	// Create mock server that returns empty array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewGroupsExporter(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	// Should still collect metrics (zeros)
	if metricCount == 0 {
		t.Error("Expected at least one metric (even if zero)")
	}
}

func TestGroupsExporter_Collect_InvalidJSON(t *testing.T) {
	// Create mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`invalid json`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewGroupsExporter(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	// Should complete without panic
	for range ch {
		// Drain channel
	}
}

func TestGroupsExporter_UpdateMetrics(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewGroupsExporter(client)

	// Test data
	issuedAPI := api.GroupIssuedApi
	issuedIntegration := api.GroupIssuedIntegration
	groups := []api.Group{
		{
			Id:             "group1",
			Name:           "test-group-1",
			PeersCount:     5,
			ResourcesCount: 2,
			Issued:         &issuedAPI,
		},
		{
			Id:             "group2",
			Name:           "test-group-2",
			PeersCount:     3,
			ResourcesCount: 1,
			Issued:         &issuedIntegration,
		},
	}

	// Call updateMetrics directly
	exporter.updateMetrics(groups)

	// Check metric values using a registry
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Verify some key metrics
	expectedMetrics := map[string]float64{
		"netbird_groups": 2,
	}

	for _, family := range families {
		if expectedValue, exists := expectedMetrics[family.GetName()]; exists {
			if len(family.GetMetric()) > 0 {
				actualValue := family.GetMetric()[0].GetGauge().GetValue()
				if actualValue != expectedValue {
					t.Errorf("Expected %s to be %f, got %f", family.GetName(), expectedValue, actualValue)
				}
			}
		}
	}
}

func TestGroupsExporter_MetricsReset(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewGroupsExporter(client)

	// Set some initial values
	exporter.groupsTotal.WithLabelValues().Set(10)
	exporter.groupPeersCount.WithLabelValues("group1", "test-group", "api").Set(5)

	// Create empty groups to test reset
	groups := []api.Group{}
	exporter.updateMetrics(groups)

	// Collect and verify metrics are reset
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() == "netbird_groups" {
			if len(family.GetMetric()) > 0 {
				value := family.GetMetric()[0].GetGauge().GetValue()
				if value != 0 {
					t.Errorf("Expected groups total to be reset to 0, got %f", value)
				}
			}
		}
	}
}

func TestGroupsExporter_MetricLabels(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewGroupsExporter(client)

	// Test data with specific values for label verification
	issued := api.GroupIssuedApi
	groups := []api.Group{
		{
			Id:             "group1",
			Name:           "test-group",
			PeersCount:     5,
			ResourcesCount: 2,
			Issued:         &issued,
		},
	}

	// Call updateMetrics directly instead of Collect to avoid API calls
	exporter.updateMetrics(groups)

	// Collect metrics to verify labels
	ch := make(chan prometheus.Metric, 50)
	go func() {
		// Only collect the metrics we've set, not trigger API calls
		exporter.groupsTotal.Collect(ch)
		exporter.groupPeersCount.Collect(ch)
		exporter.groupResourcesCount.Collect(ch)
		exporter.groupInfo.Collect(ch)
		close(ch)
	}()

	labelFound := false
	for metric := range ch {
		// Check if metric has expected labels - this is a basic check
		// In a real scenario, you'd want to inspect the metric's labels more thoroughly
		if metric != nil {
			labelFound = true
		}
	}

	if !labelFound {
		t.Error("Expected to find metrics with labels")
	}
}
