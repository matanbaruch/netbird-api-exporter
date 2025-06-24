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

func TestNewDNSExporter(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewDNSExporter(client)

	if exporter == nil {
		t.Fatal("Expected exporter to be non-nil")
	}

	if exporter.client != client {
		t.Error("Expected client to be set correctly")
	}

	// Check that all metrics are initialized
	if exporter.nameserverGroupsTotal == nil {
		t.Error("Expected nameserverGroupsTotal metric to be non-nil")
	}
	if exporter.nameserverGroupsEnabled == nil {
		t.Error("Expected nameserverGroupsEnabled metric to be non-nil")
	}
	if exporter.nameserverGroupsPrimary == nil {
		t.Error("Expected nameserverGroupsPrimary metric to be non-nil")
	}
	if exporter.nameserverGroupDomains == nil {
		t.Error("Expected nameserverGroupDomains metric to be non-nil")
	}
	if exporter.nameserversTotal == nil {
		t.Error("Expected nameserversTotal metric to be non-nil")
	}
	if exporter.nameserversByType == nil {
		t.Error("Expected nameserversByType metric to be non-nil")
	}
	if exporter.nameserversByPort == nil {
		t.Error("Expected nameserversByPort metric to be non-nil")
	}
	if exporter.dnsManagementDisabled == nil {
		t.Error("Expected dnsManagementDisabled metric to be non-nil")
	}
}

func TestDNSExporter_Describe(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewDNSExporter(client)

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

func TestDNSExporter_Collect_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Token ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/api/dns/nameservers":
			nameserverGroups := []api.NameserverGroup{
				{
					Id:          "ns1",
					Name:        "primary-dns",
					Description: "Primary DNS servers",
					Nameservers: []api.Nameserver{
						{Ip: "8.8.8.8", NsType: "udp", Port: 53},
						{Ip: "8.8.4.4", NsType: "udp", Port: 53},
					},
					Enabled:              true,
					Groups:               []string{"group1", "group2"},
					Primary:              true,
					Domains:              []string{"example.com", "test.com"},
					SearchDomainsEnabled: true,
				},
				{
					Id:          "ns2",
					Name:        "secondary-dns",
					Description: "Secondary DNS servers",
					Nameservers: []api.Nameserver{
						{Ip: "1.1.1.1", NsType: "udp", Port: 53},
					},
					Enabled:              false,
					Groups:               []string{"group3"},
					Primary:              false,
					Domains:              []string{"internal.com"},
					SearchDomainsEnabled: false,
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(nameserverGroups); err != nil {
				t.Errorf("Failed to encode nameserver groups: %v", err)
			}

		case "/api/dns/settings":
			settings := api.DNSSettings{
				DisabledManagementGroups: []string{"group4", "group5"},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(settings); err != nil {
				t.Errorf("Failed to encode settings: %v", err)
			}

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewDNSExporter(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
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

	// Check nameserver groups total
	totalFound := false
	for _, family := range families {
		if family.GetName() == "netbird_dns_nameserver_groups" {
			totalFound = true
			if len(family.GetMetric()) > 0 {
				value := family.GetMetric()[0].GetGauge().GetValue()
				if value != 2 {
					t.Errorf("Expected nameserver groups total to be 2, got %f", value)
				}
			}
			break
		}
	}

	if !totalFound {
		t.Error("Expected to find nameserver groups total metric")
	}
}

func TestDNSExporter_Collect_APIError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewDNSExporter(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
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

func TestDNSExporter_Collect_EmptyResponse(t *testing.T) {
	// Create mock server that returns empty responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/api/dns/nameservers":
			if _, err := w.Write([]byte(`[]`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		case "/api/dns/settings":
			if _, err := w.Write([]byte(`{"items": {"disabled_management_groups": []}}`)); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewDNSExporter(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
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

func TestDNSExporter_UpdateNameserverMetrics(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewDNSExporter(client)

	// Test data
	nameserverGroups := []api.NameserverGroup{
		{
			Id:          "ns1",
			Name:        "primary-dns",
			Description: "Primary DNS",
			Nameservers: []api.Nameserver{
				{Ip: "8.8.8.8", NsType: "udp", Port: 53},
				{Ip: "8.8.4.4", NsType: "udp", Port: 53},
			},
			Enabled:              true,
			Primary:              true,
			Domains:              []string{"example.com", "test.com"},
			SearchDomainsEnabled: true,
		},
		{
			Id:          "ns2",
			Name:        "secondary-dns",
			Description: "Secondary DNS",
			Nameservers: []api.Nameserver{
				{Ip: "1.1.1.1", NsType: "udp", Port: 53},
			},
			Enabled:              false,
			Primary:              false,
			Domains:              []string{"internal.com"},
			SearchDomainsEnabled: false,
		},
	}

	// Call updateNameserverMetrics directly
	exporter.updateNameserverMetrics(nameserverGroups)

	// Check metric values using a registry
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Verify some key metrics
	expectedMetrics := map[string]float64{
		"netbird_dns_nameserver_groups": 2,
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
