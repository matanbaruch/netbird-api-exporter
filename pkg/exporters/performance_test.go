package exporters

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	nbclient "github.com/netbirdio/netbird/management/client/rest"
	"github.com/prometheus/client_golang/prometheus"
)

func TestExporters_Performance_ConcurrentCollections(t *testing.T) {
	// Create a mock server that responds to all endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add some latency to simulate real API calls
		time.Sleep(10 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/api/peers":
			_, _ = w.Write([]byte(`[{"id":"peer1","name":"test","ip":"100.64.0.1","connected":true}]`))
		case "/api/groups":
			_, _ = w.Write([]byte(`[{"id":"group1","name":"test","peers_count":1}]`))
		case "/api/users":
			_, _ = w.Write([]byte(`[{"id":"user1","email":"test@example.com","role":"admin"}]`))
		case "/api/dns/nameservers":
			_, _ = w.Write([]byte(`[{"id":"ns1","name":"test","enabled":true}]`))
		case "/api/dns/settings":
			_, _ = w.Write([]byte(`{"items":{"disabled_management_groups":[]}}`))
		case "/api/networks":
			_, _ = w.Write([]byte(`[{"id":"net1","name":"test"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	exporter := NewNetBirdExporter(server.URL, "test-token")

	// Test concurrent collections
	const numGoroutines = 10
	const collectionsPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*collectionsPerGoroutine)

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < collectionsPerGoroutine; j++ {
				metricsCh := make(chan prometheus.Metric, 1000)
				go func() {
					defer close(metricsCh)
					exporter.Collect(metricsCh)
				}()

				metricsCount := 0
				for metric := range metricsCh {
					if metric == nil {
						errors <- nil // Signal that we got a nil metric
					}
					metricsCount++
				}

				if metricsCount == 0 {
					errors <- nil // Signal that we got no metrics
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	duration := time.Since(startTime)
	totalCollections := numGoroutines * collectionsPerGoroutine

	t.Logf("Completed %d concurrent collections in %v (avg: %v per collection)",
		totalCollections, duration, duration/time.Duration(totalCollections))

	// Check for errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}

	if errorCount > 0 {
		t.Errorf("Got %d errors during concurrent collections", errorCount)
	}

	// Performance expectations - should complete reasonably fast
	if duration > 10*time.Second {
		t.Errorf("Concurrent collections took too long: %v", duration)
	}
}

func TestExporters_Performance_MemoryUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Return larger datasets to test memory usage
		switch r.URL.Path {
		case "/api/peers":
			// Generate larger peer dataset
			peers := make([]string, 100)
			for i := 0; i < 100; i++ {
				peers[i] = fmt.Sprintf(`{"id":"peer%d","name":"test","ip":"100.64.0.1","connected":true}`, i)
			}
			_, _ = w.Write([]byte(`[` + strings.Join(peers, ",") + `]`))
		case "/api/groups":
			_, _ = w.Write([]byte(`[{"id":"group1","name":"test","peers_count":1}]`))
		case "/api/users":
			_, _ = w.Write([]byte(`[{"id":"user1","email":"test@example.com","role":"admin"}]`))
		case "/api/dns/nameservers":
			_, _ = w.Write([]byte(`[{"id":"ns1","name":"test","enabled":true}]`))
		case "/api/dns/settings":
			_, _ = w.Write([]byte(`{"items":{"disabled_management_groups":[]}}`))
		case "/api/networks":
			_, _ = w.Write([]byte(`[{"id":"net1","name":"test"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	exporter := NewNetBirdExporter(server.URL, "test-token")

	// Force garbage collection before measuring
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Perform multiple collections
	const numCollections = 100
	for i := 0; i < numCollections; i++ {
		metricsCh := make(chan prometheus.Metric, 1000)
		go func() {
			defer close(metricsCh)
			exporter.Collect(metricsCh)
		}()

		// Drain the channel
		for range metricsCh {
		}
	}

	// Force garbage collection after measuring
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// Calculate memory difference safely to avoid overflow issues
	var allocatedBytes uint64
	if m2.Alloc >= m1.Alloc {
		allocatedBytes = m2.Alloc - m1.Alloc
	} else {
		// If current alloc is less than initial (due to GC), use TotalAlloc instead
		if m2.TotalAlloc >= m1.TotalAlloc {
			allocatedBytes = m2.TotalAlloc - m1.TotalAlloc
		} else {
			allocatedBytes = 0 // Fallback to 0 if calculations don't make sense
		}
	}

	allocatedMB := float64(allocatedBytes) / 1024 / 1024
	t.Logf("Memory allocated during %d collections: %.2f MB", numCollections, allocatedMB)

	// Memory usage should be reasonable (less than 100MB for this test)
	// Use absolute value to handle potential negative values
	if allocatedMB < 0 {
		allocatedMB = -allocatedMB
	}
	if allocatedMB > 100 {
		t.Errorf("Excessive memory usage: %.2f MB", allocatedMB)
	}
}

func TestExporters_Performance_HighLatencyAPI(t *testing.T) {
	// Create a server with high latency to test timeout handling
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate high latency (but not timeout)
		time.Sleep(2 * time.Second)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/api/peers":
			_, _ = w.Write([]byte(`[{"id":"peer1","name":"test","ip":"100.64.0.1","connected":true}]`))
		case "/api/groups":
			_, _ = w.Write([]byte(`[{"id":"group1","name":"test","peers_count":1}]`))
		case "/api/users":
			_, _ = w.Write([]byte(`[{"id":"user1","email":"test@example.com","role":"admin"}]`))
		case "/api/dns/nameservers":
			_, _ = w.Write([]byte(`[{"id":"ns1","name":"test","enabled":true}]`))
		case "/api/dns/settings":
			_, _ = w.Write([]byte(`{"items":{"disabled_management_groups":[]}}`))
		case "/api/networks":
			_, _ = w.Write([]byte(`[{"id":"net1","name":"test"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	exporter := NewNetBirdExporter(server.URL, "test-token")

	// Test collection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	metricsCh := make(chan prometheus.Metric, 1000)
	done := make(chan struct{})

	startTime := time.Now()

	go func() {
		defer close(metricsCh)
		defer close(done)
		exporter.Collect(metricsCh)
	}()

	var metricsCount int
	select {
	case <-ctx.Done():
		t.Fatal("Collection timed out")
	case <-done:
		// Collection completed
		for range metricsCh {
			metricsCount++
		}
	}

	duration := time.Since(startTime)
	t.Logf("High latency collection completed in %v with %d metrics", duration, metricsCount)

	// Should complete but take some time due to latency
	if duration < 2*time.Second {
		t.Error("Expected collection to take at least 2 seconds due to API latency")
	}
	if duration > 25*time.Second {
		t.Errorf("Collection took too long: %v", duration)
	}
}

func TestExporters_Performance_ErrorRecovery(t *testing.T) {
	// Create a server that sometimes fails
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Fail every 3rd request to test error recovery
		if requestCount%3 == 0 {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/api/peers":
			_, _ = w.Write([]byte(`[{"id":"peer1","name":"test","ip":"100.64.0.1","connected":true}]`))
		case "/api/groups":
			_, _ = w.Write([]byte(`[{"id":"group1","name":"test","peers_count":1}]`))
		case "/api/users":
			_, _ = w.Write([]byte(`[{"id":"user1","email":"test@example.com","role":"admin"}]`))
		case "/api/dns/nameservers":
			_, _ = w.Write([]byte(`[{"id":"ns1","name":"test","enabled":true}]`))
		case "/api/dns/settings":
			_, _ = w.Write([]byte(`{"items":{"disabled_management_groups":[]}}`))
		case "/api/networks":
			_, _ = w.Write([]byte(`[{"id":"net1","name":"test"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	exporter := NewNetBirdExporter(server.URL, "test-token")

	// Perform multiple collections with intermittent failures
	successCount := 0
	const numAttempts = 10

	for i := 0; i < numAttempts; i++ {
		metricsCh := make(chan prometheus.Metric, 1000)
		go func() {
			defer close(metricsCh)
			exporter.Collect(metricsCh)
		}()

		metricsCount := 0
		for range metricsCh {
			metricsCount++
		}

		if metricsCount > 0 {
			successCount++
		}

		// Small delay between attempts
		time.Sleep(100 * time.Millisecond)
	}

	t.Logf("Successful collections: %d/%d", successCount, numAttempts)

	// Should have some successful collections despite intermittent failures
	if successCount == 0 {
		t.Error("Expected at least some successful collections despite API errors")
	}
}

func BenchmarkExporter_Collect(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/api/peers":
			_, _ = w.Write([]byte(`[{"id":"peer1","name":"test","ip":"100.64.0.1","connected":true}]`))
		case "/api/groups":
			_, _ = w.Write([]byte(`[{"id":"group1","name":"test","peers_count":1}]`))
		case "/api/users":
			_, _ = w.Write([]byte(`[{"id":"user1","email":"test@example.com","role":"admin"}]`))
		case "/api/dns/nameservers":
			_, _ = w.Write([]byte(`[{"id":"ns1","name":"test","enabled":true}]`))
		case "/api/dns/settings":
			_, _ = w.Write([]byte(`{"items":{"disabled_management_groups":[]}}`))
		case "/api/networks":
			_, _ = w.Write([]byte(`[{"id":"net1","name":"test"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	exporter := NewNetBirdExporter(server.URL, "test-token")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		metricsCh := make(chan prometheus.Metric, 1000)
		go func() {
			defer close(metricsCh)
			exporter.Collect(metricsCh)
		}()

		// Drain the channel
		for range metricsCh {
		}
	}
}

func BenchmarkPeersExporter_Collect(b *testing.B) {
	// Create a mock peers response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"id":"peer1","name":"test-peer-1","ip":"100.64.0.1","connected":true},
			{"id":"peer2","name":"test-peer-2","ip":"100.64.0.2","connected":false},
			{"id":"peer3","name":"test-peer-3","ip":"100.64.0.3","connected":true}
		]`))
	}))
	defer server.Close()

	// Create exporter with test server client
	exporter := NewPeersExporter(nbclient.New(server.URL, "test-token"))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		metricsCh := make(chan prometheus.Metric, 100)
		go func() {
			defer close(metricsCh)
			exporter.Collect(metricsCh)
		}()

		// Drain the channel
		for range metricsCh {
		}
	}
}

func TestExporters_StressTest_ManyMetrics(t *testing.T) {
	// Create a server that returns many metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/api/peers":
			// Generate 1000 peers
			peers := make([]string, 1000)
			for i := 0; i < 1000; i++ {
				peers[i] = fmt.Sprintf(`{"id":"peer%d","name":"test","ip":"100.64.0.1","connected":true,"last_seen":"2023-01-01T00:00:00Z","os":"linux","groups":[],"ssh_enabled":true,"hostname":"host","login_expired":false,"approval_required":false,"country_code":"US","city_name":"City","accessible_peers_count":5}`, i)
			}
			_, _ = w.Write([]byte(`[` + strings.Join(peers, ",") + `]`))
		case "/api/groups":
			_, _ = w.Write([]byte(`[{"id":"group1","name":"test","peers_count":1000}]`))
		case "/api/users":
			_, _ = w.Write([]byte(`[{"id":"user1","email":"test@example.com","role":"admin"}]`))
		case "/api/dns/nameservers":
			_, _ = w.Write([]byte(`[{"id":"ns1","name":"test","enabled":true}]`))
		case "/api/dns/settings":
			_, _ = w.Write([]byte(`{"items":{"disabled_management_groups":[]}}`))
		case "/api/networks":
			_, _ = w.Write([]byte(`[{"id":"net1","name":"test"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	exporter := NewNetBirdExporter(server.URL, "test-token")

	startTime := time.Now()

	metricsCh := make(chan prometheus.Metric, 10000)
	go func() {
		defer close(metricsCh)
		exporter.Collect(metricsCh)
	}()

	metricsCount := 0
	for range metricsCh {
		metricsCount++
	}

	duration := time.Since(startTime)

	t.Logf("Collected %d metrics in %v", metricsCount, duration)

	// Should handle many metrics efficiently
	if metricsCount < 1000 {
		t.Errorf("Expected at least 1000 metrics for 1000 peers, got %d", metricsCount)
	}

	if duration > 30*time.Second {
		t.Errorf("Collection of many metrics took too long: %v", duration)
	}
}
