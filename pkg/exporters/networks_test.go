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

func TestNewNetworksExporter(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewNetworksExporter(client)

	if exporter == nil {
		t.Fatal("Expected exporter to be non-nil")
	}

	if exporter.client != client {
		t.Error("Expected client to be set correctly")
	}

	// Check that all metrics are initialized
	if exporter.networksTotal == nil {
		t.Error("Expected networksTotal metric to be non-nil")
	}
	if exporter.networkRoutersCount == nil {
		t.Error("Expected networkRoutersCount metric to be non-nil")
	}
	if exporter.networkResourcesCount == nil {
		t.Error("Expected networkResourcesCount metric to be non-nil")
	}
	if exporter.networkPoliciesCount == nil {
		t.Error("Expected networkPoliciesCount metric to be non-nil")
	}
	if exporter.networkRoutingPeersCount == nil {
		t.Error("Expected networkRoutingPeersCount metric to be non-nil")
	}
	if exporter.networkInfo == nil {
		t.Error("Expected networkInfo metric to be non-nil")
	}
}

func TestNetworksExporter_Describe(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewNetworksExporter(client)

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

func TestNetworksExporter_Collect_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/networks" {
			http.NotFound(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Token ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		description1 := "Production network"
		description2 := "Development network"
		networks := []api.Network{
			{
				Id:                "net1",
				Name:              "production-network",
				Description:       &description1,
				Routers:           []string{"router1", "router2"},
				RoutingPeersCount: 5,
				Resources:         []string{"resource1", "resource2", "resource3"},
				Policies:          []string{"policy1"},
			},
			{
				Id:                "net2",
				Name:              "development-network",
				Description:       &description2,
				Routers:           []string{"router3"},
				RoutingPeersCount: 2,
				Resources:         []string{"resource4"},
				Policies:          []string{"policy2", "policy3"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(networks); err != nil {
			t.Errorf("Failed to encode networks: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewNetworksExporter(client)

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

	// Check networks total
	totalFound := false
	for _, family := range families {
		if family.GetName() == "netbird_networks" {
			totalFound = true
			if len(family.GetMetric()) > 0 {
				value := family.GetMetric()[0].GetGauge().GetValue()
				if value != 2 {
					t.Errorf("Expected networks total to be 2, got %f", value)
				}
			}
			break
		}
	}

	if !totalFound {
		t.Error("Expected to find networks total metric")
	}
}

func TestNetworksExporter_Collect_APIError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewNetworksExporter(client)

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

func TestNetworksExporter_Collect_EmptyResponse(t *testing.T) {
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
	exporter := NewNetworksExporter(client)

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

func TestNetworksExporter_Collect_InvalidJSON(t *testing.T) {
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
	exporter := NewNetworksExporter(client)

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

func TestNetworksExporter_UpdateMetrics(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewNetworksExporter(client)

	// Test data
	description1 := "Test network 1"
	description2 := "Test network 2"
	networks := []api.Network{
		{
			Id:                "net1",
			Name:              "test-network-1",
			Description:       &description1,
			Routers:           []string{"router1", "router2"},
			RoutingPeersCount: 5,
			Resources:         []string{"resource1", "resource2"},
			Policies:          []string{"policy1"},
		},
		{
			Id:                "net2",
			Name:              "test-network-2",
			Description:       &description2,
			Routers:           []string{"router3"},
			RoutingPeersCount: 3,
			Resources:         []string{"resource3"},
			Policies:          []string{"policy2", "policy3"},
		},
	}

	// Call updateMetrics directly
	exporter.updateMetrics(networks)

	// Check metric values using a registry
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Verify some key metrics
	expectedMetrics := map[string]float64{
		"netbird_networks": 2,
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

func TestNetworksExporter_MetricsReset(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewNetworksExporter(client)

	// Set some initial values
	exporter.networksTotal.WithLabelValues().Set(10)
	exporter.networkRoutersCount.WithLabelValues("net1", "test-network").Set(5)

	// Create empty networks to test reset
	networks := []api.Network{}
	exporter.updateMetrics(networks)

	// Collect and verify metrics are reset
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() == "netbird_networks" {
			if len(family.GetMetric()) > 0 {
				value := family.GetMetric()[0].GetGauge().GetValue()
				if value != 0 {
					t.Errorf("Expected networks total to be reset to 0, got %f", value)
				}
			}
		}
	}
}
