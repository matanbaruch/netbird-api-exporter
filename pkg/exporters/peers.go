package exporters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"netbird-api-exporter/pkg/netbird"
)

// PeersExporter handles peers-specific metrics collection
type PeersExporter struct {
	client *netbird.Client

	// Prometheus metrics
	peersTotal                 *prometheus.GaugeVec
	peersConnected             *prometheus.GaugeVec
	peersLastSeen              *prometheus.GaugeVec
	peersByOS                  *prometheus.GaugeVec
	peersByCountry             *prometheus.GaugeVec
	peersByGroup               *prometheus.GaugeVec
	peersSSHEnabled            *prometheus.GaugeVec
	peersLoginExpired          *prometheus.GaugeVec
	peersApprovalRequired      *prometheus.GaugeVec
	accessiblePeersCount       *prometheus.GaugeVec
	peerConnectionStatusByName *prometheus.GaugeVec
}

// NewPeersExporter creates a new peers exporter
func NewPeersExporter(client *netbird.Client) *PeersExporter {
	return &PeersExporter{
		client: client,

		peersTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peers",
				Help: "Total number of NetBird peers",
			},
			[]string{},
		),

		peersConnected: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peers_connected",
				Help: "Number of connected NetBird peers",
			},
			[]string{"connected"},
		),

		peersLastSeen: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peer_last_seen_timestamp",
				Help: "Last seen timestamp of NetBird peers",
			},
			[]string{"peer_id", "peer_name", "hostname"},
		),

		peersByOS: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peers_by_os",
				Help: "Number of NetBird peers by operating system",
			},
			[]string{"os"},
		),

		peersByCountry: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peers_by_country",
				Help: "Number of NetBird peers by country",
			},
			[]string{"country_code", "city_name"},
		),

		peersByGroup: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peers_by_group",
				Help: "Number of NetBird peers by group",
			},
			[]string{"group_id", "group_name"},
		),

		peersSSHEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peers_ssh_enabled",
				Help: "Number of NetBird peers with SSH enabled",
			},
			[]string{"ssh_enabled"},
		),

		peersLoginExpired: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peers_login_expired",
				Help: "Number of NetBird peers with expired login",
			},
			[]string{"login_expired"},
		),

		peersApprovalRequired: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peers_approval_required",
				Help: "Number of NetBird peers requiring approval",
			},
			[]string{"approval_required"},
		),

		accessiblePeersCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peer_accessible_peers_count",
				Help: "Number of accessible peers for each peer",
			},
			[]string{"peer_id", "peer_name"},
		),

		peerConnectionStatusByName: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netbird_peer_connection_status_by_name",
				Help: "Connection status of each peer by name (1 for connected, 0 for disconnected)",
			},
			[]string{"peer_name", "peer_id", "connected"},
		),
	}
}

// Describe implements prometheus.Collector
func (e *PeersExporter) Describe(ch chan<- *prometheus.Desc) {
	e.peersTotal.Describe(ch)
	e.peersConnected.Describe(ch)
	e.peersLastSeen.Describe(ch)
	e.peersByOS.Describe(ch)
	e.peersByCountry.Describe(ch)
	e.peersByGroup.Describe(ch)
	e.peersSSHEnabled.Describe(ch)
	e.peersLoginExpired.Describe(ch)
	e.peersApprovalRequired.Describe(ch)
	e.accessiblePeersCount.Describe(ch)
	e.peerConnectionStatusByName.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *PeersExporter) Collect(ch chan<- prometheus.Metric) {
	// Reset metrics before collecting new values
	e.peersTotal.Reset()
	e.peersConnected.Reset()
	e.peersLastSeen.Reset()
	e.peersByOS.Reset()
	e.peersByCountry.Reset()
	e.peersByGroup.Reset()
	e.peersSSHEnabled.Reset()
	e.peersLoginExpired.Reset()
	e.peersApprovalRequired.Reset()
	e.accessiblePeersCount.Reset()
	e.peerConnectionStatusByName.Reset()

	peers, err := e.fetchPeers()
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch peers")
		return
	}

	e.updateMetrics(peers)

	// Collect all metrics
	e.peersTotal.Collect(ch)
	e.peersConnected.Collect(ch)
	e.peersLastSeen.Collect(ch)
	e.peersByOS.Collect(ch)
	e.peersByCountry.Collect(ch)
	e.peersByGroup.Collect(ch)
	e.peersSSHEnabled.Collect(ch)
	e.peersLoginExpired.Collect(ch)
	e.peersApprovalRequired.Collect(ch)
	e.accessiblePeersCount.Collect(ch)
	e.peerConnectionStatusByName.Collect(ch)
}

// fetchPeers retrieves peers from NetBird API
func (e *PeersExporter) fetchPeers() ([]netbird.Peer, error) {
	url := fmt.Sprintf("%s/api/peers", e.client.GetBaseURL())

	logrus.WithField("url", url).Debug("Fetching peers from NetBird API")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", e.client.GetToken()))

	resp, err := e.client.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logrus.WithError(closeErr).Warn("Failed to close response body for peers")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var peers []netbird.Peer
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	logrus.WithField("count", len(peers)).Debug("Successfully fetched peers from API")

	return peers, nil
}

// updateMetrics updates Prometheus metrics based on peer data
func (e *PeersExporter) updateMetrics(peers []netbird.Peer) {
	// Count totals
	totalPeers := len(peers)
	connectedCount := 0
	disconnectedCount := 0

	// Count by categories
	osCounts := make(map[string]int)
	countryCounts := make(map[string]int)
	groupCounts := make(map[string]int)
	sshEnabledCount := 0
	sshDisabledCount := 0
	loginExpiredCount := 0
	loginValidCount := 0
	approvalRequiredCount := 0
	approvalNotRequiredCount := 0

	for _, peer := range peers {
		// Connection status
		if peer.Connected {
			connectedCount++
		} else {
			disconnectedCount++
		}

		// Last seen timestamp
		e.peersLastSeen.WithLabelValues(peer.ID, peer.Name, peer.Hostname).Set(float64(peer.LastSeen.Unix()))

		// OS distribution
		osKey := peer.OS
		if osKey == "" {
			osKey = "unknown"
		}
		osCounts[osKey]++

		// Country distribution
		countryKey := fmt.Sprintf("%s_%s", peer.CountryCode, peer.CityName)
		if peer.CountryCode == "" {
			countryKey = "unknown_unknown"
		}
		countryCounts[countryKey]++

		// Group membership
		for _, group := range peer.Groups {
			groupKey := fmt.Sprintf("%s_%s", group.ID, group.Name)
			groupCounts[groupKey]++
		}

		// SSH status
		if peer.SSHEnabled {
			sshEnabledCount++
		} else {
			sshDisabledCount++
		}

		// Login status
		if peer.LoginExpired {
			loginExpiredCount++
		} else {
			loginValidCount++
		}

		// Approval status
		if peer.ApprovalRequired {
			approvalRequiredCount++
		} else {
			approvalNotRequiredCount++
		}

		// Accessible peers count
		e.accessiblePeersCount.WithLabelValues(peer.ID, peer.Name).Set(float64(peer.AccessiblePeersCount))

		// Connection status by name - using peer.Name for peer_name label
		connectedStr := "false"
		connectionValue := 0.0
		if peer.Connected {
			connectedStr = "true"
			connectionValue = 1.0
		}
		e.peerConnectionStatusByName.WithLabelValues(peer.Name, peer.ID, connectedStr).Set(connectionValue)
	}

	// Set metrics
	e.peersTotal.WithLabelValues().Set(float64(totalPeers))
	e.peersConnected.WithLabelValues("true").Set(float64(connectedCount))
	e.peersConnected.WithLabelValues("false").Set(float64(disconnectedCount))

	// OS distribution
	for os, count := range osCounts {
		e.peersByOS.WithLabelValues(os).Set(float64(count))
	}

	// Country distribution
	for countryCity, count := range countryCounts {
		parts := strings.SplitN(countryCity, "_", 2)
		country := parts[0]
		city := ""
		if len(parts) > 1 {
			city = parts[1]
		}
		e.peersByCountry.WithLabelValues(country, city).Set(float64(count))
	}

	// Group distribution
	for groupInfo, count := range groupCounts {
		parts := strings.SplitN(groupInfo, "_", 2)
		groupID := parts[0]
		groupName := ""
		if len(parts) > 1 {
			groupName = parts[1]
		}
		e.peersByGroup.WithLabelValues(groupID, groupName).Set(float64(count))
	}

	// SSH status
	e.peersSSHEnabled.WithLabelValues("true").Set(float64(sshEnabledCount))
	e.peersSSHEnabled.WithLabelValues("false").Set(float64(sshDisabledCount))

	// Login status
	e.peersLoginExpired.WithLabelValues("true").Set(float64(loginExpiredCount))
	e.peersLoginExpired.WithLabelValues("false").Set(float64(loginValidCount))

	// Approval status
	e.peersApprovalRequired.WithLabelValues("true").Set(float64(approvalRequiredCount))
	e.peersApprovalRequired.WithLabelValues("false").Set(float64(approvalNotRequiredCount))

	logrus.WithFields(logrus.Fields{
		"total_peers":             totalPeers,
		"connected_peers":         connectedCount,
		"disconnected_peers":      disconnectedCount,
		"ssh_enabled_peers":       sshEnabledCount,
		"ssh_disabled_peers":      sshDisabledCount,
		"login_expired_peers":     loginExpiredCount,
		"login_valid_peers":       loginValidCount,
		"approval_required_peers": approvalRequiredCount,
		"os_distributions":        len(osCounts),
		"country_distributions":   len(countryCounts),
		"group_memberships":       len(groupCounts),
	}).Debug("Updated peer metrics")
}
