package exporters

import (
	"context"
	"time"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// AccountsExporter handles accounts-specific metrics collection
type AccountsExporter struct {
	client *nbclient.Client

	// Prometheus metrics for accounts
	accountInfo        *prometheus.GaugeVec
	accountCreatedAt   *prometheus.GaugeVec
	scrapeErrorsTotal  *prometheus.CounterVec
	scrapeDuration     *prometheus.HistogramVec
}

// NewAccountsExporter creates a new accounts exporter
func NewAccountsExporter(client *nbclient.Client) *AccountsExporter {
	return &AccountsExporter{
		client: client,

		accountInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_account_info",
				Help: "Information about the NetBird account (always 1)",
			},
			[]string{"account_id", "domain", "domain_category", "created_by"},
		),

		accountCreatedAt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_account_created_at_timestamp",
				Help: "Unix timestamp when the account was created",
			},
			[]string{"account_id", "domain"},
		),

		scrapeErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netbird_accounts_scrape_errors_total",
				Help: "Total number of errors encountered while scraping accounts",
			},
			[]string{"error_type"},
		),

		scrapeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "netbird_accounts_scrape_duration_seconds",
				Help: "Time spent scraping accounts from the NetBird API",
			},
			[]string{},
		),
	}
}

// Describe implements prometheus.Collector
func (e *AccountsExporter) Describe(ch chan<- *prometheus.Desc) {
	e.accountInfo.Describe(ch)
	e.accountCreatedAt.Describe(ch)
	e.scrapeErrorsTotal.Describe(ch)
	e.scrapeDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *AccountsExporter) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(e.scrapeDuration.WithLabelValues())
	defer timer.ObserveDuration()

	// Reset metrics before collecting new values
	e.accountInfo.Reset()
	e.accountCreatedAt.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	accounts, err := e.client.Accounts.List(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch accounts")
		e.scrapeErrorsTotal.WithLabelValues("fetch_accounts").Inc()
		return
	}

	e.updateMetrics(accounts)

	// Collect all metrics
	e.accountInfo.Collect(ch)
	e.accountCreatedAt.Collect(ch)
	e.scrapeErrorsTotal.Collect(ch)
	e.scrapeDuration.Collect(ch)
}

// updateMetrics updates Prometheus metrics based on accounts data
func (e *AccountsExporter) updateMetrics(accounts []api.Account) {
	for _, account := range accounts {
		labels := []string{account.Id, account.Domain, account.DomainCategory, account.CreatedBy}
		e.accountInfo.WithLabelValues(labels...).Set(1)

		timestampLabels := []string{account.Id, account.Domain}
		e.accountCreatedAt.WithLabelValues(timestampLabels...).Set(float64(account.CreatedAt.Unix()))
	}

	logrus.WithFields(logrus.Fields{
		"total_accounts": len(accounts),
	}).Debug("Updated account metrics")
}
