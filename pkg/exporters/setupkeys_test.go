package exporters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewSetupKeysExporter(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewSetupKeysExporter(client)

	if exporter == nil {
		t.Fatal("Expected exporter to be non-nil")
	}

	if exporter.client != client {
		t.Error("Expected client to be set correctly")
	}

	if exporter.setupKeysTotal == nil {
		t.Error("Expected setupKeysTotal metric to be non-nil")
	}
	if exporter.setupKeysValid == nil {
		t.Error("Expected setupKeysValid metric to be non-nil")
	}
	if exporter.setupKeyUsedTimes == nil {
		t.Error("Expected setupKeyUsedTimes metric to be non-nil")
	}
	if exporter.setupKeyInfo == nil {
		t.Error("Expected setupKeyInfo metric to be non-nil")
	}
}

func TestSetupKeysExporter_Describe(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewSetupKeysExporter(client)

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

func testSetupKeys() []api.SetupKey {
	return []api.SetupKey{
		{
			Id:         "key1",
			Name:       "reusable-key",
			Type:       "reusable",
			State:      "valid",
			Valid:      true,
			Revoked:    false,
			Ephemeral:  false,
			UsedTimes:  3,
			UsageLimit: 0,
			Expires:    time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUsed:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			AutoGroups: []string{"group1", "group2"},
		},
		{
			Id:         "key2",
			Name:       "one-off-key",
			Type:       "one-off",
			State:      "expired",
			Valid:      false,
			Revoked:    false,
			Ephemeral:  true,
			UsedTimes:  1,
			UsageLimit: 1,
			Expires:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUsed:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			AutoGroups: []string{},
		},
	}
}

func TestSetupKeysExporter_Collect_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/setup-keys" {
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
		if err := json.NewEncoder(w).Encode(testSetupKeys()); err != nil {
			t.Errorf("Failed to encode setup keys: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewSetupKeysExporter(client)

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	validFound := false
	for _, family := range families {
		if family.GetName() == "netbird_setup_keys_valid" {
			validFound = true
			for _, m := range family.GetMetric() {
				for _, l := range m.GetLabel() {
					if l.GetName() == "valid" && l.GetValue() == "true" {
						if m.GetGauge().GetValue() != 1 {
							t.Errorf("Expected 1 valid setup key, got %f", m.GetGauge().GetValue())
						}
					}
				}
			}
		}
	}

	if !validFound {
		t.Error("Expected to find netbird_setup_keys_valid metric")
	}
}

func TestSetupKeysExporter_Collect_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewSetupKeysExporter(client)

	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	for range ch {
		// Drain channel; must not panic or hang
	}
}

func TestSetupKeysExporter_Collect_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewSetupKeysExporter(client)

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

func TestSetupKeysExporter_Collect_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`invalid json`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewSetupKeysExporter(client)

	ch := make(chan prometheus.Metric, 50)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	for range ch {
		// Drain channel
	}
}

func TestSetupKeysExporter_UpdateMetrics(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewSetupKeysExporter(client)

	exporter.updateMetrics(testSetupKeys())

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		exporter.setupKeysTotal,
		exporter.setupKeysValid,
		exporter.setupKeysRevoked,
		exporter.setupKeysEphemeral,
		exporter.setupKeyUsedTimes,
		exporter.setupKeyUsageLimit,
		exporter.setupKeyAutoGroups,
	)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() == "netbird_setup_key_used_times" {
			for _, m := range family.GetMetric() {
				var keyID string
				for _, l := range m.GetLabel() {
					if l.GetName() == "key_id" {
						keyID = l.GetValue()
					}
				}
				if keyID == "key1" && m.GetGauge().GetValue() != 3 {
					t.Errorf("Expected key1 used_times to be 3, got %f", m.GetGauge().GetValue())
				}
			}
		}
	}
}

func TestSetupKeysExporter_MetricsReset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewSetupKeysExporter(client)

	// Set a stale series that should be cleared on the next collection
	exporter.setupKeyUsedTimes.WithLabelValues("key1", "stale-key").Set(99)

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter.setupKeyUsedTimes)

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
		if family.GetName() == "netbird_setup_key_used_times" {
			if len(family.GetMetric()) != 0 {
				t.Errorf("Expected setup key used_times to be reset, got %d metrics", len(family.GetMetric()))
			}
		}
	}
}
