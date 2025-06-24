package pkg

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	nbclient "github.com/netbirdio/netbird/management/client/rest"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/matanbaruch/netbird-api-exporter/pkg/exporters"
	"github.com/matanbaruch/netbird-api-exporter/pkg/utils"
)

// Integration tests that require a real NetBird API token
// These tests are skipped if NETBIRD_API_TOKEN is not set

func TestIntegration_SkipIfNoToken(t *testing.T) {
	token := os.Getenv("NETBIRD_API_TOKEN")
	if token == "" {
		t.Skip("Skipping integration tests: NETBIRD_API_TOKEN environment variable not set")
	}
}

func getTestClient(t *testing.T) *nbclient.Client {
	t.Helper()

	url, token := getURLAndCreds(t)
	client := nbclient.New(url, token)

	if client == nil {
		t.Fatal("Failed to create NetBird client")
	}

	return client
}

func getURLAndCreds(t *testing.T) (string, string) {
	t.Helper()
	baseURL := utils.GetEnvWithDefault("NETBIRD_API_URL", "https://api.netbird.io")
	token := os.Getenv("NETBIRD_API_TOKEN")
	if token == "" {
		t.Skipf("NETBIRD_API_TOKEN environment variable not set")
	}
	return baseURL, token
}

func TestIntegration_NetBirdExporter_RealAPI(t *testing.T) {
	token := os.Getenv("NETBIRD_API_TOKEN")
	if token == "" {
		t.Skip("Skipping integration test: NETBIRD_API_TOKEN environment variable not set")
	}

	baseURL := utils.GetEnvWithDefault("NETBIRD_API_URL", "https://api.netbird.io")
	exporter := exporters.NewNetBirdExporter(baseURL, token)

	// Set up Prometheus registry for testing
	registry := prometheus.NewRegistry()
	if err := registry.Register(exporter); err != nil {
		t.Fatalf("Failed to register exporter: %v", err)
	}

	// Test that Describe works
	ch := make(chan *prometheus.Desc, 100)
	go func() {
		exporter.Describe(ch)
		close(ch)
	}()

	descCount := 0
	for desc := range ch {
		if desc == nil {
			t.Error("Received nil metric description")
		}
		descCount++
	}

	if descCount == 0 {
		t.Error("Expected at least one metric description")
	}

	// Test that Collect works with real API
	metricsCh := make(chan prometheus.Metric, 1000)
	go func() {
		defer close(metricsCh)

		// Add timeout to prevent hanging
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			defer close(done)
			exporter.Collect(metricsCh)
		}()

		select {
		case <-ctx.Done():
			t.Error("Collect operation timed out")
			return
		case <-done:
			// Collection completed successfully
		}
	}()

	metricsCount := 0
	var metrics []prometheus.Metric
	for metric := range metricsCh {
		if metric == nil {
			t.Error("Received nil metric")
			continue
		}
		metrics = append(metrics, metric)
		metricsCount++
	}

	if metricsCount == 0 {
		t.Error("Expected at least one metric from real API")
	}

	t.Logf("Successfully collected %d metrics from NetBird API", metricsCount)

	// Verify we get expected metric families
	expectedMetricPrefixes := []string{
		"netbird_peers_",
		"netbird_groups_",
		"netbird_users_",
		"netbird_dns_",
		"netbird_networks_",
		"netbird_exporter_scrape_",
	}

	metricNames := make(map[string]bool)
	for _, metric := range metrics {
		desc := metric.Desc()
		if desc != nil {
			// Extract metric name from descriptor
			descStr := desc.String()
			// Parse the fqName from the descriptor string
			// Format: Desc{fqName: "metric_name", help: "...", ...}
			start := strings.Index(descStr, `fqName: "`) + 9
			if start > 8 {
				end := strings.Index(descStr[start:], `"`)
				if end > 0 {
					metricName := descStr[start : start+end]
					metricNames[metricName] = true
				}
			}
		}
	}

	for _, prefix := range expectedMetricPrefixes {
		found := false
		for name := range metricNames {
			if strings.HasPrefix(name, prefix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("No metrics found with prefix %s", prefix)
		}
	}
}

func TestIntegration_PeersExporter_RealData(t *testing.T) {
	client := getTestClient(t)

	peersExporter := exporters.NewPeersExporter(client)
	if peersExporter == nil {
		t.Fatal("Failed to create peers exporter")
	}

	// Test metrics collection
	metricsCh := make(chan prometheus.Metric, 100)
	go func() {
		defer close(metricsCh)
		peersExporter.Collect(metricsCh)
	}()

	var metrics []prometheus.Metric
	for metric := range metricsCh {
		metrics = append(metrics, metric)
	}

	// Should have at least the basic metrics even if no peers exist
	if len(metrics) == 0 {
		t.Error("Expected at least some peer metrics")
	}

	t.Logf("Collected %d peer metrics", len(metrics))
}

func TestIntegration_GroupsExporter_RealData(t *testing.T) {
	client := getTestClient(t)

	groupsExporter := exporters.NewGroupsExporter(client)
	if groupsExporter == nil {
		t.Fatal("Failed to create groups exporter")
	}

	metricsCh := make(chan prometheus.Metric, 100)
	go func() {
		defer close(metricsCh)
		groupsExporter.Collect(metricsCh)
	}()

	var metrics []prometheus.Metric
	for metric := range metricsCh {
		metrics = append(metrics, metric)
	}

	if len(metrics) == 0 {
		t.Error("Expected at least some group metrics")
	}

	t.Logf("Collected %d group metrics", len(metrics))
}

func TestIntegration_UsersExporter_RealData(t *testing.T) {
	client := getTestClient(t)

	usersExporter := exporters.NewUsersExporter(client)
	if usersExporter == nil {
		t.Fatal("Failed to create users exporter")
	}

	metricsCh := make(chan prometheus.Metric, 100)
	go func() {
		defer close(metricsCh)
		usersExporter.Collect(metricsCh)
	}()

	var metrics []prometheus.Metric
	for metric := range metricsCh {
		metrics = append(metrics, metric)
	}

	if len(metrics) == 0 {
		t.Error("Expected at least some user metrics")
	}

	t.Logf("Collected %d user metrics", len(metrics))
}

func TestIntegration_DNSExporter_RealData(t *testing.T) {
	client := getTestClient(t)

	dnsExporter := exporters.NewDNSExporter(client)
	if dnsExporter == nil {
		t.Fatal("Failed to create DNS exporter")
	}

	metricsCh := make(chan prometheus.Metric, 100)
	go func() {
		defer close(metricsCh)
		dnsExporter.Collect(metricsCh)
	}()

	var metrics []prometheus.Metric
	for metric := range metricsCh {
		metrics = append(metrics, metric)
	}

	if len(metrics) == 0 {
		t.Error("Expected at least some DNS metrics")
	}

	t.Logf("Collected %d DNS metrics", len(metrics))
}

func TestIntegration_NetworksExporter_RealData(t *testing.T) {
	client := getTestClient(t)

	networksExporter := exporters.NewNetworksExporter(client)
	if networksExporter == nil {
		t.Fatal("Failed to create networks exporter")
	}

	metricsCh := make(chan prometheus.Metric, 100)
	go func() {
		defer close(metricsCh)
		networksExporter.Collect(metricsCh)
	}()

	var metrics []prometheus.Metric
	for metric := range metricsCh {
		metrics = append(metrics, metric)
	}

	if len(metrics) == 0 {
		t.Error("Expected at least some network metrics")
	}

	t.Logf("Collected %d network metrics", len(metrics))
}

func TestIntegration_APIRateLimiting(t *testing.T) {
	client := getTestClient(t)

	// Test multiple rapid requests to ensure rate limiting doesn't break
	const numRequests = 5
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for i := 0; i < numRequests; i++ {
		_, err := client.Users.List(ctx)
		if err != nil {
			t.Errorf("Request %d failed: %v", i+1, err)
		}
		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}
}

func TestIntegration_MetricsAccuracy(t *testing.T) {
	// Test that metrics are consistent between multiple collections
	exporter := exporters.NewNetBirdExporter(getURLAndCreds(t))

	// Collect metrics twice
	var firstCollection, secondCollection []prometheus.Metric

	// First collection
	metricsCh1 := make(chan prometheus.Metric, 1000)
	go func() {
		defer close(metricsCh1)
		exporter.Collect(metricsCh1)
	}()

	for metric := range metricsCh1 {
		firstCollection = append(firstCollection, metric)
	}

	// Wait a bit between collections
	time.Sleep(1 * time.Second)

	// Second collection
	metricsCh2 := make(chan prometheus.Metric, 1000)
	go func() {
		defer close(metricsCh2)
		exporter.Collect(metricsCh2)
	}()

	for metric := range metricsCh2 {
		secondCollection = append(secondCollection, metric)
	}

	// Both collections should have metrics
	if len(firstCollection) == 0 {
		t.Error("First collection returned no metrics")
	}
	if len(secondCollection) == 0 {
		t.Error("Second collection returned no metrics")
	}

	// The number of metrics should be similar (allowing for some variation due to timing)
	ratio := float64(len(secondCollection)) / float64(len(firstCollection))
	if ratio < 0.8 || ratio > 1.2 {
		t.Errorf("Metric counts vary too much between collections: %d vs %d (ratio: %.2f)",
			len(firstCollection), len(secondCollection), ratio)
	}

	t.Logf("First collection: %d metrics, Second collection: %d metrics",
		len(firstCollection), len(secondCollection))
}
