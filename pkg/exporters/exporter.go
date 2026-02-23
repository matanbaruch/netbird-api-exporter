package exporters

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	nbclient "github.com/netbirdio/netbird/shared/management/client/rest"
)

// NetBirdExporter represents the main Prometheus exporter for NetBird APIs
type NetBirdExporter struct {
	client               *nbclient.Client
	peersExporter        *PeersExporter
	groupsExporter       *GroupsExporter
	usersExporter        *UsersExporter
	dnsExporter          *DNSExporter
	networksExporter     *NetworksExporter
	accountsExporter     *AccountsExporter
	setupKeysExporter    *SetupKeysExporter
	policiesExporter     *PoliciesExporter
	routesExporter       *RoutesExporter
	postureChecksExporter *PostureChecksExporter
	dnsZonesExporter     *DNSZonesExporter
	tokensExporter       *TokensExporter
	eventsExporter       *EventsExporter

	// Common metrics
	scrapeDuration prometheus.Histogram
	scrapeErrors   prometheus.Counter
}

// NewNetBirdExporter creates a new NetBird exporter with all sub-exporters
func NewNetBirdExporter(baseURL, token string) *NetBirdExporter {

	client := nbclient.New(baseURL, token)

	return &NetBirdExporter{
		client:                client,
		peersExporter:         NewPeersExporter(client),
		groupsExporter:        NewGroupsExporter(client),
		usersExporter:         NewUsersExporter(client),
		dnsExporter:           NewDNSExporter(client),
		networksExporter:      NewNetworksExporter(client),
		accountsExporter:      NewAccountsExporter(client),
		setupKeysExporter:     NewSetupKeysExporter(client),
		policiesExporter:      NewPoliciesExporter(client),
		routesExporter:        NewRoutesExporter(client),
		postureChecksExporter: NewPostureChecksExporter(client),
		dnsZonesExporter:      NewDNSZonesExporter(client),
		tokensExporter:        NewTokensExporter(client),
		eventsExporter:        NewEventsExporter(client),

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
}

// Describe implements prometheus.Collector
func (e *NetBirdExporter) Describe(ch chan<- *prometheus.Desc) {
	e.peersExporter.Describe(ch)
	e.groupsExporter.Describe(ch)
	e.usersExporter.Describe(ch)
	e.dnsExporter.Describe(ch)
	e.networksExporter.Describe(ch)
	e.accountsExporter.Describe(ch)
	e.setupKeysExporter.Describe(ch)
	e.policiesExporter.Describe(ch)
	e.routesExporter.Describe(ch)
	e.postureChecksExporter.Describe(ch)
	e.dnsZonesExporter.Describe(ch)
	e.tokensExporter.Describe(ch)
	e.eventsExporter.Describe(ch)
	e.scrapeDuration.Describe(ch)
	e.scrapeErrors.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *NetBirdExporter) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		e.scrapeDuration.Observe(duration.Seconds())
		e.scrapeDuration.Collect(ch)
		e.scrapeErrors.Collect(ch)
		logrus.WithField("total_duration", duration).Debug("Completed NetBird metrics collection")
	}()

	logrus.Debug("Starting NetBird metrics collection")

	// Collect from all sub-exporters
	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during peers collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting peers collection")
		e.peersExporter.Collect(ch)
		logrus.Debug("Completed peers collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during groups collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting groups collection")
		e.groupsExporter.Collect(ch)
		logrus.Debug("Completed groups collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during users collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting users collection")
		e.usersExporter.Collect(ch)
		logrus.Debug("Completed users collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during dns collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting DNS collection")
		e.dnsExporter.Collect(ch)
		logrus.Debug("Completed DNS collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during networks collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting networks collection")
		e.networksExporter.Collect(ch)
		logrus.Debug("Completed networks collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during accounts collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting accounts collection")
		e.accountsExporter.Collect(ch)
		logrus.Debug("Completed accounts collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during setup keys collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting setup keys collection")
		e.setupKeysExporter.Collect(ch)
		logrus.Debug("Completed setup keys collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during policies collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting policies collection")
		e.policiesExporter.Collect(ch)
		logrus.Debug("Completed policies collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during routes collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting routes collection")
		e.routesExporter.Collect(ch)
		logrus.Debug("Completed routes collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during posture checks collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting posture checks collection")
		e.postureChecksExporter.Collect(ch)
		logrus.Debug("Completed posture checks collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during DNS zones collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting DNS zones collection")
		e.dnsZonesExporter.Collect(ch)
		logrus.Debug("Completed DNS zones collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during tokens collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting tokens collection")
		e.tokensExporter.Collect(ch)
		logrus.Debug("Completed tokens collection")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic during events collection")
				e.scrapeErrors.Inc()
			}
		}()
		logrus.Debug("Starting events collection")
		e.eventsExporter.Collect(ch)
		logrus.Debug("Completed events collection")
	}()
}
