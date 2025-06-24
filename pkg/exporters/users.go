package exporters

import (
	"context"
	"time"

	nbclient "github.com/netbirdio/netbird/management/client/rest"
	"github.com/netbirdio/netbird/management/server/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// UsersExporter handles users-specific metrics collection
type UsersExporter struct {
	client *nbclient.Client

	// Prometheus metrics for users
	usersTotal           *prometheus.GaugeVec
	usersByRole          *prometheus.GaugeVec
	usersByStatus        *prometheus.GaugeVec
	usersServiceUsers    *prometheus.GaugeVec
	usersBlocked         *prometheus.GaugeVec
	usersByIssued        *prometheus.GaugeVec
	usersLastLogin       *prometheus.GaugeVec
	usersAutoGroupsCount *prometheus.GaugeVec
	usersRestricted      *prometheus.GaugeVec
	usersPermissions     *prometheus.GaugeVec
	scrapeErrorsTotal    *prometheus.CounterVec
	scrapeDuration       *prometheus.HistogramVec
}

// NewUsersExporter creates a new users exporter
func NewUsersExporter(client *nbclient.Client) *UsersExporter {
	return &UsersExporter{
		client: client,

		usersTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_users",
				Help: "Total number of NetBird users",
			},
			[]string{},
		),

		usersByRole: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_users_by_role",
				Help: "Number of NetBird users by role",
			},
			[]string{"role"},
		),

		usersByStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_users_by_status",
				Help: "Number of NetBird users by status",
			},
			[]string{"status"},
		),

		usersServiceUsers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_users_service_users",
				Help: "Number of NetBird service users vs regular users",
			},
			[]string{"is_service_user"},
		),

		usersBlocked: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_users_blocked",
				Help: "Number of blocked NetBird users",
			},
			[]string{"is_blocked"},
		),

		usersByIssued: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_users_by_issued",
				Help: "Number of NetBird users by issuance type",
			},
			[]string{"issued"},
		),

		usersLastLogin: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_user_last_login_timestamp",
				Help: "Last login timestamp of NetBird users",
			},
			[]string{"user_id", "user_email", "user_name"},
		),

		usersAutoGroupsCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_user_auto_groups_count",
				Help: "Number of auto groups assigned to each NetBird user",
			},
			[]string{"user_id", "user_email", "user_name"},
		),

		usersRestricted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_users_restricted",
				Help: "Number of NetBird users with restricted permissions",
			},
			[]string{"is_restricted"},
		),

		usersPermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_user_permissions",
				Help: "User permissions by module and action",
			},
			[]string{"user_id", "user_email", "module", "permission", "value"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_users_scrape_errors_total",
				Help: "Total number of errors encountered while scraping users",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_users_scrape_duration_seconds",
				Help: "Time spent scraping users from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *UsersExporter) Describe(ch chan<- *prometheus.Desc) {
	e.usersTotal.Describe(ch)
	e.usersByRole.Describe(ch)
	e.usersByStatus.Describe(ch)
	e.usersServiceUsers.Describe(ch)
	e.usersBlocked.Describe(ch)
	e.usersByIssued.Describe(ch)
	e.usersLastLogin.Describe(ch)
	e.usersAutoGroupsCount.Describe(ch)
	e.usersRestricted.Describe(ch)
	e.usersPermissions.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *UsersExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.usersTotal.Reset()
	e.usersByRole.Reset()
	e.usersByStatus.Reset()
	e.usersServiceUsers.Reset()
	e.usersBlocked.Reset()
	e.usersByIssued.Reset()
	e.usersLastLogin.Reset()
	e.usersAutoGroupsCount.Reset()
	e.usersRestricted.Reset()
	e.usersPermissions.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := e.client.Users.List(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch users")
		e.scrapeErrorsTotal.WithLabelValues("fetch_users").Inc()
		return
	}

	e.updateMetrics(users)

	// Collect all metrics
	e.usersTotal.Collect(ch)
	e.usersByRole.Collect(ch)
	e.usersByStatus.Collect(ch)
	e.usersServiceUsers.Collect(ch)
	e.usersBlocked.Collect(ch)
	e.usersByIssued.Collect(ch)
	e.usersLastLogin.Collect(ch)
	e.usersAutoGroupsCount.Collect(ch)
	e.usersRestricted.Collect(ch)
	e.usersPermissions.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on users data
func (e *UsersExporter) updateMetrics(users []api.User) {
	totalUsers := len(users)

	// Count by categories
	roleCounts := make(map[string]int)
	statusCounts := make(map[string]int)
	serviceUserCount := 0
	regularUserCount := 0
	blockedCount := 0
	unblockedCount := 0
	issuedCounts := make(map[string]int)
	restrictedCount := 0
	unrestrictedCount := 0
	totalPermissionsCount := 0

	for _, user := range users {
		// Role distribution
		role := user.Role
		if role == "" {
			role = "unknown"
		}
		roleCounts[role]++

		// Status distribution
		status := user.Status
		if status == "" {
			status = "unknown"
		}
		statusCounts[string(status)]++

		// Service user classification
		if user.IsServiceUser != nil && *user.IsServiceUser {
			serviceUserCount++
		} else {
			regularUserCount++
		}

		// Blocked status
		if user.IsBlocked {
			blockedCount++
		} else {
			unblockedCount++
		}

		// Issued type distribution
		issued := "unknown"
		if user.Issued != nil {
			issued = *user.Issued
		}
		issuedCounts[issued]++

		// Restricted permissions
		if user.Permissions != nil && user.Permissions.IsRestricted {
			restrictedCount++
		} else {
			unrestrictedCount++
		}

		// Individual user metrics
		userLabels := []string{user.Id, user.Email, user.Name}

		// Last login timestamp
		if user.LastLogin != nil && !user.LastLogin.IsZero() {
			e.usersLastLogin.WithLabelValues(userLabels...).Set(float64(user.LastLogin.Unix()))
		}

		// Auto groups count
		e.usersAutoGroupsCount.WithLabelValues(userLabels...).Set(float64(len(user.AutoGroups)))
		if user.Permissions == nil {
			user.Permissions = &api.UserPermissions{
				IsRestricted: false,
				Modules:      make(map[string]map[string]bool),
			}
		}
		// User permissions per module and action
		for module, permissions := range user.Permissions.Modules {
			for permission, value := range permissions {
				valueStr := "false"
				if value {
					valueStr = "true"
				}
				e.usersPermissions.WithLabelValues(user.Id, user.Email, module, permission, valueStr).Set(1)
				totalPermissionsCount++
			}
		}
	}

	// Set aggregate metrics
	e.usersTotal.WithLabelValues().Set(float64(totalUsers))

	// Role distribution
	for role, count := range roleCounts {
		e.usersByRole.WithLabelValues(role).Set(float64(count))
	}

	// Status distribution
	for status, count := range statusCounts {
		e.usersByStatus.WithLabelValues(status).Set(float64(count))
	}

	// Service user counts
	e.usersServiceUsers.WithLabelValues("true").Set(float64(serviceUserCount))
	e.usersServiceUsers.WithLabelValues("false").Set(float64(regularUserCount))

	// Blocked counts
	e.usersBlocked.WithLabelValues("true").Set(float64(blockedCount))
	e.usersBlocked.WithLabelValues("false").Set(float64(unblockedCount))

	// Issued type distribution
	for issued, count := range issuedCounts {
		e.usersByIssued.WithLabelValues(issued).Set(float64(count))
	}

	// Restricted permission counts
	e.usersRestricted.WithLabelValues("true").Set(float64(restrictedCount))
	e.usersRestricted.WithLabelValues("false").Set(float64(unrestrictedCount))

	logrus.WithFields(logrus.Fields{
		"total_users":             totalUsers,
		"service_users":           serviceUserCount,
		"regular_users":           regularUserCount,
		"blocked_users":           blockedCount,
		"unblocked_users":         unblockedCount,
		"restricted_users":        restrictedCount,
		"unrestricted_users":      unrestrictedCount,
		"role_distributions":      roleCounts,
		"status_distributions":    statusCounts,
		"issued_distributions":    issuedCounts,
		"total_permissions_count": totalPermissionsCount,
	}).Debug("Updated user metrics")
}
