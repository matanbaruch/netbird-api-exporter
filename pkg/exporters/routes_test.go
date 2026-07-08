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

func TestNewRoutesExporter(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewRoutesExporter(client)

	if exporter == nil {
		t.Fatal("Expected exporter to be non-nil")
	}

	if exporter.client != client {
		t.Error("Expected client to be set correctly")
	}

	if exporter.routesTotal == nil {
		t.Error("Expected routesTotal metric to be non-nil")
	}
	if exporter.routesByNetworkType == nil {
		t.Error("Expected routesByNetworkType metric to be non-nil")
	}
	if exporter.routesMasquerade == nil {
		t.Error("Expected routesMasquerade metric to be non-nil")
	}
	if exporter.routeInfo == nil {
		t.Error("Expected routeInfo metric to be non-nil")
	}
}

func TestRoutesExporter_Describe(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewRoutesExporter(client)

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

func testRoutes() []api.Route {
	network := "10.0.0.0/24"
	return []api.Route{
		{
			Id:          "route1",
			NetworkId:   "net-a",
			NetworkType: "IPv4",
			Description: "internal network",
			Enabled:     true,
			Masquerade:  true,
			Metric:      9999,
			Network:     &network,
			Groups:      []string{"group1"},
		},
		{
			Id:          "route2",
			NetworkId:   "net-b",
			NetworkType: "Domain",
			Description: "domain route",
			Enabled:     false,
			Masquerade:  false,
			Metric:      100,
			Domains:     &[]string{"example.com"},
			Groups:      []string{"group2"},
		},
	}
}

func TestRoutesExporter_Collect_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/routes" {
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
		if err := json.NewEncoder(w).Encode(testRoutes()); err != nil {
			t.Errorf("Failed to encode routes: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewRoutesExporter(client)

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	typeFound := false
	for _, family := range families {
		if family.GetName() == "netbird_routes_by_network_type" {
			typeFound = true
			for _, m := range family.GetMetric() {
				for _, l := range m.GetLabel() {
					if l.GetName() == "network_type" && l.GetValue() == "IPv4" {
						if m.GetGauge().GetValue() != 1 {
							t.Errorf("Expected 1 IPv4 route, got %f", m.GetGauge().GetValue())
						}
					}
				}
			}
		}
	}

	if !typeFound {
		t.Error("Expected to find netbird_routes_by_network_type metric")
	}
}

func TestRoutesExporter_Collect_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewRoutesExporter(client)

	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	for range ch {
		// Drain channel; must not panic or hang
	}
}

func TestRoutesExporter_Collect_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewRoutesExporter(client)

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

func TestRoutesExporter_Collect_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`invalid json`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewRoutesExporter(client)

	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	for range ch {
		// Drain channel
	}
}

func TestRoutesExporter_UpdateMetrics(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewRoutesExporter(client)

	exporter.updateMetrics(testRoutes())

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		exporter.routesTotal,
		exporter.routesByNetworkType,
		exporter.routesMasquerade,
		exporter.routeInfo,
	)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() == "netbird_routes" {
			for _, m := range family.GetMetric() {
				for _, l := range m.GetLabel() {
					if l.GetName() == "enabled" && l.GetValue() == "true" {
						if m.GetGauge().GetValue() != 1 {
							t.Errorf("Expected 1 enabled route, got %f", m.GetGauge().GetValue())
						}
					}
				}
			}
		}
	}
}

func TestRoutesExporter_MetricsReset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewRoutesExporter(client)

	// Set a stale series that should be cleared on the next collection
	exporter.routeInfo.WithLabelValues("route1", "net-a", "IPv4", "stale").Set(1)

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter.routeInfo)

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
		if family.GetName() == "netbird_route_info" {
			if len(family.GetMetric()) != 0 {
				t.Errorf("Expected route info to be reset, got %d metrics", len(family.GetMetric()))
			}
		}
	}
}
