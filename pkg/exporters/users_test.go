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
	"github.com/netbirdio/netbird/util"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewUsersExporter(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewUsersExporter(client)

	if exporter == nil {
		t.Fatal("Expected exporter to be non-nil")
	}

	if exporter.client != client {
		t.Error("Expected client to be set correctly")
	}

	// Check that all metrics are initialized
	metrics := map[string]interface{}{
		"usersTotal":           exporter.usersTotal,
		"usersByRole":          exporter.usersByRole,
		"usersByStatus":        exporter.usersByStatus,
		"usersServiceUsers":    exporter.usersServiceUsers,
		"usersBlocked":         exporter.usersBlocked,
		"usersByIssued":        exporter.usersByIssued,
		"usersLastLogin":       exporter.usersLastLogin,
		"usersAutoGroupsCount": exporter.usersAutoGroupsCount,
		"usersRestricted":      exporter.usersRestricted,
		"usersPermissions":     exporter.usersPermissions,
		"scrapeErrorsTotal":    exporter.scrapeErrorsTotal,
		"scrapeDuration":       exporter.scrapeDuration,
	}

	for name, metric := range metrics {
		if metric == nil {
			t.Errorf("Expected %s metric to be non-nil", name)
		}
	}
}

func TestUsersExporter_Describe(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewUsersExporter(client)

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

func TestUsersExporter_Collect_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users" {
			http.NotFound(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Token ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		lastTimeNow := time.Now()
		lastTime24 := lastTimeNow.Add(-time.Hour * 24)
		lastTime1 := lastTimeNow.Add(-time.Hour * 1)
		issuedAPI := "api"
		issuedIntegration := "integration"
		users := []api.User{
			{
				Id:            "user1",
				Email:         "admin@example.com",
				Name:          "Admin User",
				Role:          "admin",
				Status:        "active",
				LastLogin:     &lastTimeNow,
				AutoGroups:    []string{"group1", "group2"},
				IsCurrent:     util.False(),
				IsServiceUser: util.False(),
				IsBlocked:     false,
				Issued:        &issuedAPI,
				Permissions: &api.UserPermissions{
					IsRestricted: false,
					Modules: map[string]map[string]bool{
						"peers": {
							"read":   true,
							"write":  true,
							"delete": true,
						},
					},
				},
			},
			{
				Id:            "user2",
				Email:         "service@example.com",
				Name:          "Service User",
				Role:          "user",
				Status:        "active",
				LastLogin:     &lastTime1,
				AutoGroups:    []string{"group1"},
				IsCurrent:     util.False(),
				IsServiceUser: util.True(),
				IsBlocked:     false,
				Issued:        &issuedIntegration,
				Permissions: &api.UserPermissions{
					IsRestricted: true,
					Modules: map[string]map[string]bool{
						"peers": {
							"read":   true,
							"write":  false,
							"delete": false,
						},
					},
				},
			},
			{
				Id:            "user3",
				Email:         "blocked@example.com",
				Name:          "Blocked User",
				Role:          "user",
				Status:        "inactive",
				LastLogin:     &lastTime24,
				AutoGroups:    []string{},
				IsCurrent:     util.False(),
				IsServiceUser: util.False(),
				IsBlocked:     true,
				Issued:        &issuedAPI,
				Permissions: &api.UserPermissions{
					IsRestricted: false,
					Modules:      map[string]map[string]bool{},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(users); err != nil {
			t.Errorf("Failed to encode users: %v", err)
		}
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewUsersExporter(client)

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

	// Check users total
	totalFound := false
	for _, family := range families {
		if family.GetName() == "netbird_users" {
			totalFound = true
			if len(family.GetMetric()) > 0 {
				value := family.GetMetric()[0].GetGauge().GetValue()
				if value != 3 {
					t.Errorf("Expected users total to be 3, got %f", value)
				}
			}
			break
		}
	}

	if !totalFound {
		t.Error("Expected to find users total metric")
	}
}

func TestUsersExporter_Collect_APIError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := nbclient.New(server.URL, "test-token")
	exporter := NewUsersExporter(client)

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

	// Even with API errors, some metrics might still be collected (like error counters)
	// This test ensures the collection doesn't panic or hang
}

func TestUsersExporter_Collect_EmptyResponse(t *testing.T) {
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
	exporter := NewUsersExporter(client)

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

func TestUsersExporter_UpdateMetrics(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewUsersExporter(client)

	// Test data
	issuedAPI := "api"
	issuedIntegration := "integration"
	lastLogin := time.Now()
	users := []api.User{
		{
			Id:            "user1",
			Email:         "admin@example.com",
			Name:          "Admin User",
			Role:          "admin",
			Status:        "active",
			LastLogin:     &lastLogin,
			AutoGroups:    []string{"group1", "group2"},
			IsServiceUser: util.False(),
			IsBlocked:     false,
			Issued:        &issuedAPI,
			Permissions: &api.UserPermissions{
				IsRestricted: false,
				Modules: map[string]map[string]bool{
					"peers": {
						"read":  true,
						"write": true,
					},
				},
			},
		},
		{
			Id:            "user2",
			Email:         "user@example.com",
			Name:          "Regular User",
			Role:          "user",
			Status:        "active",
			IsServiceUser: util.True(),
			IsBlocked:     false,
			Issued:        &issuedIntegration,
			AutoGroups:    []string{"group1"},
			Permissions: &api.UserPermissions{
				IsRestricted: true,
				Modules:      map[string]map[string]bool{},
			},
		},
	}

	// Call updateMetrics directly
	exporter.updateMetrics(users)

	// Check metric values using a registry
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Verify some key metrics
	expectedMetrics := map[string]float64{
		"netbird_users": 2,
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

func TestUsersExporter_MetricsReset(t *testing.T) {
	client := nbclient.New("https://api.netbird.io", "test-token")
	exporter := NewUsersExporter(client)

	// Set some initial values
	exporter.usersTotal.WithLabelValues().Set(10)
	exporter.usersByRole.WithLabelValues("admin").Set(5)

	// Create empty users to test reset
	users := []api.User{}
	exporter.updateMetrics(users)

	// Collect and verify metrics are reset
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() == "netbird_users" {
			if len(family.GetMetric()) > 0 {
				value := family.GetMetric()[0].GetGauge().GetValue()
				if value != 0 {
					t.Errorf("Expected users total to be reset to 0, got %f", value)
				}
			}
		}
	}
}
