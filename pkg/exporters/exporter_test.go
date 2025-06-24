package exporters

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	nbclient "github.com/netbirdio/netbird/management/client/rest"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewNetBirdExporter(t *testing.T) {
	baseURL := "https://api.netbird.io"
	token := "test-token"

	exporter := NewNetBirdExporter(baseURL, token)

	if exporter == nil {
		t.Fatal("Expected exporter to be non-nil")
	}

	if exporter.client == nil {
		t.Error("Expected client to be non-nil")
	}

	if exporter.peersExporter == nil {
		t.Error("Expected peersExporter to be non-nil")
	}

	if exporter.groupsExporter == nil {
		t.Error("Expected groupsExporter to be non-nil")
	}

	if exporter.usersExporter == nil {
		t.Error("Expected usersExporter to be non-nil")
	}

	if exporter.dnsExporter == nil {
		t.Error("Expected dnsExporter to be non-nil")
	}

	if exporter.networksExporter == nil {
		t.Error("Expected networksExporter to be non-nil")
	}

	if exporter.scrapeDuration == nil {
		t.Error("Expected scrapeDuration metric to be non-nil")
	}

	if exporter.scrapeErrors == nil {
		t.Error("Expected scrapeErrors metric to be non-nil")
	}
}

func TestNetBirdExporter_Describe(t *testing.T) {
	exporter := NewNetBirdExporter("https://api.netbird.io", "test-token")

	ch := make(chan *prometheus.Desc, 100)
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

func TestNetBirdExporter_Collect_WithMockServer(t *testing.T) {
	// Create mock server for testing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Token ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/api/peers":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`[
				{
					"id": "peer1",
					"name": "test-peer",
					"ip": "100.64.0.1",
					"connected": true,
					"last_seen": "2023-01-01T00:00:00Z",
					"os": "linux",
					"groups": [],
					"ssh_enabled": true,
					"hostname": "test-host",
					"login_expired": false,
					"approval_required": false,
					"country_code": "US",
					"city_name": "New York",
					"accessible_peers_count": 5
				}
			]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/groups":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`[
				{
					"id": "group1",
					"name": "test-group",
					"peers_count": 1,
					"resources_count": 0,
					"issued": "api"
				}
			]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/users":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`[
				{
					"id": "user1",
					"email": "test@example.com",
					"name": "Test User",
					"role": "admin",
					"status": "active",
					"last_login": "2023-01-01T00:00:00Z",
					"auto_groups": [],
					"is_service_user": false,
					"is_blocked": false,
					"issued": "api",
					"permissions": {
						"is_restricted": false,
						"modules": {}
					}
				}
			]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/dns/nameservers":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`[
				{
					"id": "ns1",
					"name": "test-ns",
					"description": "Test nameserver",
					"nameservers": [{"ip": "8.8.8.8", "ns_type": "udp", "port": 53}],
					"enabled": true,
					"groups": [],
					"primary": true,
					"domains": ["example.com"],
					"search_domains_enabled": true
				}
			]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/dns/settings":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{
				"items": {
					"disabled_management_groups": []
				}
			}`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/networks":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`[
				{
					"id": "net1",
					"name": "test-network",
					"description": "Test network",
					"routers": [],
					"routing_peers_count": 0,
					"resources": [],
					"policies": []
				}
			]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	exporter := NewNetBirdExporter(server.URL, "test-token")

	// Test collection
	ch := make(chan prometheus.Metric, 100)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	metricCount := 0
	for metric := range ch {
		if metric == nil {
			t.Error("Expected metric to be non-nil")
		}
		metricCount++
	}

	if metricCount == 0 {
		t.Error("Expected at least one metric to be collected")
	}
}

func TestNetBirdExporter_Collect_HandlesErrors(t *testing.T) {
	// Create mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	exporter := NewNetBirdExporter(server.URL, "test-token")

	// Test collection with errors
	ch := make(chan prometheus.Metric, 100)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	// Should still collect some metrics (even if just error counters)
	metricCount := 0
	for metric := range ch {
		if metric == nil {
			t.Error("Expected metric to be non-nil")
		}
		metricCount++
	}

	// Should have at least the duration and error metrics
	if metricCount == 0 {
		t.Error("Expected at least error metrics to be collected")
	}
}

func TestNetBirdExporter_Collect_HandlesPanics(t *testing.T) {
	invalidClient := nbclient.New("http://invalid", "token")
	// Create an exporter with a nil client to potentially cause panics
	exporter := &NetBirdExporter{
		client:           nil,
		peersExporter:    NewPeersExporter(invalidClient),
		groupsExporter:   NewGroupsExporter(invalidClient),
		usersExporter:    NewUsersExporter(invalidClient),
		dnsExporter:      NewDNSExporter(invalidClient),
		networksExporter: NewNetworksExporter(invalidClient),
		scrapeDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name: "netbird_exporter_scrape_duration_seconds",
				Help: "Time spent scraping NetBird API",
			},
		),
		scrapeErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "netbird_exporter_scrape_errors_total",
				Help: "Total number of scrape errors",
			},
		),
	}

	// This should not panic even if individual exporters fail
	ch := make(chan prometheus.Metric, 100)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Collect should not panic, but got: %v", r)
			}
			close(ch)
		}()
		exporter.Collect(ch)
	}()

	// Wait for collection to complete
	for range ch {
		// Drain the channel
	}
}

func TestNetBirdExporter_MetricsRegistration(t *testing.T) {
	registry := prometheus.NewRegistry()
	exporter := NewNetBirdExporter("https://api.netbird.io", "test-token")

	err := registry.Register(exporter)
	if err != nil {
		t.Fatalf("Failed to register exporter: %v", err)
	}

	// Test that metrics can be gathered
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	if len(families) == 0 {
		t.Error("Expected at least one metric family")
	}
}

func TestNetBirdExporter_ScrapeMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/api/peers":
			if _, err := w.Write([]byte(`[]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/groups":
			if _, err := w.Write([]byte(`[]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/users":
			if _, err := w.Write([]byte(`[]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/dns/nameservers":
			if _, err := w.Write([]byte(`[]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/dns/settings":
			if _, err := w.Write([]byte(`{"items": {"disabled_management_groups": []}}`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/networks":
			if _, err := w.Write([]byte(`[]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	exporter := NewNetBirdExporter(server.URL, "test-token")
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	// Collect metrics
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Check for scrape duration metric
	found := false
	for _, family := range families {
		if family.GetName() == "netbird_exporter_scrape_duration_seconds" {
			found = true
			if len(family.GetMetric()) == 0 {
				t.Error("Expected scrape duration metric to have values")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find scrape duration metric")
	}
}
