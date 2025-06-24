package exporters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	nbclient "github.com/netbirdio/netbird/management/client/rest"
	"github.com/netbirdio/netbird/management/server/http/api"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewPeersExporter(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewPeersExporter(client)

	if exporter == nil {
		t.Fatal("Expected exporter to be non-nil")
	}

	if exporter.client != client {
		t.Error("Expected client to be set correctly")
	}

	// Check that all metrics are initialized
	if exporter.peersTotal == nil {
		t.Error("Expected peersTotal metric to be non-nil")
	}
	if exporter.peersConnected == nil {
		t.Error("Expected peersConnected metric to be non-nil")
	}
	if exporter.peersLastSeen == nil {
		t.Error("Expected peersLastSeen metric to be non-nil")
	}
	if exporter.peersByOS == nil {
		t.Error("Expected peersByOS metric to be non-nil")
	}
	if exporter.peersByCountry == nil {
		t.Error("Expected peersByCountry metric to be non-nil")
	}
	if exporter.peersByGroup == nil {
		t.Error("Expected peersByGroup metric to be non-nil")
	}
	if exporter.peersSSHEnabled == nil {
		t.Error("Expected peersSSHEnabled metric to be non-nil")
	}
	if exporter.peersLoginExpired == nil {
		t.Error("Expected peersLoginExpired metric to be non-nil")
	}
	if exporter.peersApprovalRequired == nil {
		t.Error("Expected peersApprovalRequired metric to be non-nil")
	}
	if exporter.accessiblePeersCount == nil {
		t.Error("Expected accessiblePeersCount metric to be non-nil")
	}
	if exporter.peerConnectionStatusByName == nil {
		t.Error("Expected peerConnectionStatusByName metric to be non-nil")
	}
}

func TestPeersExporter_Describe(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewPeersExporter(client)

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

func TestPeersExporter_Collect_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/peers" {
			http.NotFound(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Token ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		peers := []api.Peer{
			{
				Id:               "peer1",
				Name:             "test-peer-1",
				Ip:               "100.64.0.1",
				Connected:        true,
				LastSeen:         time.Now(),
				Os:               "linux",
				Groups:           []api.GroupMinimum{{Id: "group1", Name: "test-group"}},
				SshEnabled:       true,
				Hostname:         "test-host-1",
				LoginExpired:     false,
				ApprovalRequired: false,
				CountryCode:      "US",
				CityName:         "New York",
			},
			{
				Id:               "peer2",
				Name:             "test-peer-2",
				Ip:               "100.64.0.2",
				Connected:        false,
				LastSeen:         time.Now().Add(-time.Hour),
				Os:               "windows",
				Groups:           []api.GroupMinimum{{Id: "group2", Name: "another-group"}},
				SshEnabled:       false,
				Hostname:         "test-host-2",
				LoginExpired:     true,
				ApprovalRequired: true,
				CountryCode:      "CA",
				CityName:         "Toronto",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(peers); err != nil {
			t.Errorf("Failed to encode peers: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewPeersExporter(client)

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

	// Check peers total
	totalFound := false
	for _, family := range families {
		if family.GetName() == "netbird_peers" {
			totalFound = true
			if len(family.GetMetric()) > 0 {
				value := family.GetMetric()[0].GetGauge().GetValue()
				if value != 2 {
					t.Errorf("Expected peers total to be 2, got %f", value)
				}
			}
			break
		}
	}

	if !totalFound {
		t.Error("Expected to find peers total metric")
	}
}

func TestPeersExporter_Collect_APIError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewPeersExporter(client)

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

	// Even with API errors, some metrics might still be collected (like error counters)
	// This test ensures the collection doesn't panic or hang
}

func TestPeersExporter_Collect_InvalidJSON(t *testing.T) {
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
	exporter := NewPeersExporter(client)

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

func TestPeersExporter_Collect_EmptyResponse(t *testing.T) {
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
	exporter := NewPeersExporter(client)

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

func TestPeersExporter_Collect_Unauthorized(t *testing.T) {
	// Create mock server that requires authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		validToken := "Token valid-token" // #nosec G101 -- This is a test token, not a real credential
		if token != validToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewPeersExporter(client)

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

func TestPeersExporter_UpdateMetrics(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewPeersExporter(client)

	// Test data
	peers := []api.Peer{
		{
			Id:               "peer1",
			Name:             "test-peer-1",
			Connected:        true,
			Os:               "linux",
			Groups:           []api.GroupMinimum{{Id: "group1", Name: "test-group"}},
			SshEnabled:       true,
			LoginExpired:     false,
			ApprovalRequired: false,
			CountryCode:      "US",
			CityName:         "New York",
		},
		{
			Id:               "peer2",
			Name:             "test-peer-2",
			Connected:        false,
			Os:               "windows",
			Groups:           []api.GroupMinimum{{Id: "group1", Name: "test-group"}},
			SshEnabled:       false,
			LoginExpired:     true,
			ApprovalRequired: true,
			CountryCode:      "US",
			CityName:         "Los Angeles",
		},
	}

	// Call updateMetrics directly
	exporter.updateMetrics(peers)

	// Check metric values using a registry
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Verify some key metrics
	expectedMetrics := map[string]float64{
		"netbird_peers": 2,
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

func TestPeersExporter_MetricLabels(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewPeersExporter(client)

	// Test data with specific values for label verification
	peers := []api.Peer{
		{
			Id:          "peer1",
			Name:        "test-peer",
			Connected:   true,
			Os:          "linux",
			CountryCode: "US",
			CityName:    "New York",
			Groups:      []api.GroupMinimum{{Id: "group1", Name: "test-group"}},
		},
	}

	// Call updateMetrics directly to avoid API calls
	exporter.updateMetrics(peers)

	// Collect metrics to verify labels
	ch := make(chan prometheus.Metric, 50)
	go func() {
		// Only collect the metrics we've set, not trigger API calls
		exporter.peersTotal.Collect(ch)
		exporter.peersConnected.Collect(ch)
		exporter.peersByOS.Collect(ch)
		exporter.peersByCountry.Collect(ch)
		exporter.peersByGroup.Collect(ch)
		exporter.peersSSHEnabled.Collect(ch)
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

func TestPeersExporter_ConnectionStatusByName(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewPeersExporter(client)

	// Test data with both connected and disconnected peers
	peers := []api.Peer{
		{
			Id:        "peer1",
			Name:      "connected-peer",
			Connected: true,
		},
		{
			Id:        "peer2",
			Name:      "disconnected-peer",
			Connected: false,
		},
	}

	// Call updateMetrics directly
	exporter.updateMetrics(peers)

	// Collect just our specific metric to avoid API calls
	ch := make(chan prometheus.Metric, 10)
	go func() {
		exporter.peerConnectionStatusByName.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for metric := range ch {
		if metric != nil {
			metrics = append(metrics, metric)
		}
	}

	// Verify we have 2 metrics (one for each peer)
	if len(metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(metrics))
	}

	// Since we can't easily extract labels from prometheus.Metric interface in tests,
	// let's test by checking that we set the metrics correctly by using a test registry
	testRegistry := prometheus.NewPedanticRegistry()
	testRegistry.MustRegister(exporter.peerConnectionStatusByName)

	families, err := testRegistry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	if len(families) != 1 {
		t.Fatalf("Expected 1 metric family, got %d", len(families))
	}

	family := families[0]
	if family.GetName() != "netbird_peer_connection_status_by_name" {
		t.Errorf("Expected metric name 'netbird_peer_connection_status_by_name', got '%s'", family.GetName())
	}

	// Verify we have 2 metrics (one for each peer)
	if len(family.GetMetric()) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(family.GetMetric()))
	}

	// Verify the metrics have correct values and labels
	foundConnected := false
	foundDisconnected := false

	for _, metric := range family.GetMetric() {
		value := metric.GetGauge().GetValue()
		labels := metric.GetLabel()

		// Find peer_name and connected labels
		var peerName, connected string
		for _, label := range labels {
			if label.GetName() == "peer_name" {
				peerName = label.GetValue()
			}
			if label.GetName() == "connected" {
				connected = label.GetValue()
			}
		}

		if peerName == "connected-peer" {
			foundConnected = true
			if connected != "true" {
				t.Errorf("Expected connected label to be 'true' for connected peer, got '%s'", connected)
			}
			if value != 1.0 {
				t.Errorf("Expected value 1.0 for connected peer, got %f", value)
			}
		}

		if peerName == "disconnected-peer" {
			foundDisconnected = true
			if connected != "false" {
				t.Errorf("Expected connected label to be 'false' for disconnected peer, got '%s'", connected)
			}
			if value != 0.0 {
				t.Errorf("Expected value 0.0 for disconnected peer, got %f", value)
			}
		}
	}

	if !foundConnected {
		t.Error("Expected to find connected peer metric")
	}
	if !foundDisconnected {
		t.Error("Expected to find disconnected peer metric")
	}
}
