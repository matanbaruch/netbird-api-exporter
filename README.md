# NetBird API Exporter

![assets_task_01jwh7hmm6f93ab38cy6rb9szb_1748630092_img_0](https://github.com/user-attachments/assets/df57ed5f-524a-4965-9b8a-a8cb97ee4892)

<!-- Build and Quality Badges -->

[![Lint](https://github.com/matanbaruch/netbird-api-exporter/actions/workflows/lint.yml/badge.svg)](https://github.com/matanbaruch/netbird-api-exporter/actions/workflows/lint.yml)
[![Release](https://github.com/matanbaruch/netbird-api-exporter/actions/workflows/release.yml/badge.svg)](https://github.com/matanbaruch/netbird-api-exporter/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/matanbaruch/netbird-api-exporter)](https://goreportcard.com/report/github.com/matanbaruch/netbird-api-exporter)

<!-- Language and Tech Stack -->

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=for-the-badge&logo=Prometheus&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)
![Kubernetes](https://img.shields.io/badge/kubernetes-%23326ce5.svg?style=for-the-badge&logo=kubernetes&logoColor=white)

<!-- Version and Distribution -->

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/matanbaruch/netbird-api-exporter)](https://github.com/matanbaruch/netbird-api-exporter/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/matanbaruch/netbird-api-exporter)](https://github.com/matanbaruch/netbird-api-exporter/blob/main/go.mod)
[![License](https://img.shields.io/github/license/matanbaruch/netbird-api-exporter)](https://github.com/matanbaruch/netbird-api-exporter/blob/main/LICENSE)

<!-- GitHub Stats -->

[![GitHub stars](https://img.shields.io/github/stars/matanbaruch/netbird-api-exporter?style=social)](https://github.com/matanbaruch/netbird-api-exporter/stargazers)

<!-- Distribution Platforms -->

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/netbird-api-exporter)](https://artifacthub.io/packages/search?repo=netbird-api-exporter)

A Prometheus exporter for NetBird API that provides comprehensive metrics about your NetBird deployment. This exporter fetches data from the NetBird REST API including:

- [Peers API](https://docs.netbird.io/api/resources/peers) - Network peer metrics
- [Groups API](https://docs.netbird.io/api/resources/groups) - Group membership
- [Users API](https://docs.netbird.io/api/resources/users) - User management
- [Networks API](https://docs.netbird.io/api/resources/networks) - Network configuration
- [DNS API](https://docs.netbird.io/api/resources/dns) - DNS settings
- [Accounts API](https://docs.netbird.io/api/resources/accounts) - Account information
- [Setup Keys API](https://docs.netbird.io/api/resources/setup-keys) - Device onboarding
- [Policies API](https://docs.netbird.io/api/resources/policies) - Access policies
- [Routes API](https://docs.netbird.io/api/resources/routes) - Network routing
- [Posture Checks API](https://docs.netbird.io/api/resources/posture-checks) - Device compliance
- [DNS Zones API](https://docs.netbird.io/api/resources/dns-zones) - DNS zone management
- [Tokens API](https://docs.netbird.io/api/resources/tokens) - Personal access tokens (opt-in)
- [Events API](https://docs.netbird.io/api/resources/events) - Audit logs (opt-in)

## Metrics Overview

The exporter provides the following metrics:

### Peer Metrics

| Metric Name                              | Type  | Description                                                                  | Labels                              |
| ---------------------------------------- | ----- | ---------------------------------------------------------------------------- | ----------------------------------- |
| `netbird_peers`                          | Gauge | Total number of NetBird peers                                                | -                                   |
| `netbird_peers_connected`                | Gauge | Number of connected/disconnected peers                                       | `connected`                         |
| `netbird_peer_last_seen_timestamp`       | Gauge | Last seen timestamp for each peer                                            | `peer_id`, `peer_name`, `hostname`, `user_id`  |
| `netbird_peers_by_os`                    | Gauge | Number of peers by operating system                                          | `os`                                |
| `netbird_peers_by_country`               | Gauge | Number of peers by country/city                                              | `country_code`, `city_name`         |
| `netbird_peers_by_group`                 | Gauge | Number of peers by group                                                     | `group_id`, `group_name`            |
| `netbird_peers_ssh_enabled`              | Gauge | Number of peers with SSH enabled/disabled                                    | `ssh_enabled`                       |
| `netbird_peers_login_expired`            | Gauge | Number of peers with expired/valid login                                     | `login_expired`                     |
| `netbird_peers_approval_required`        | Gauge | Number of peers requiring/not requiring approval                             | `approval_required`                 |
| `netbird_peer_accessible_peers_count`    | Gauge | Number of accessible peers for each peer                                     | `peer_id`, `peer_name`              |
| `netbird_peer_connection_status_by_name` | Gauge | Connection status of each peer by name (1 for connected, 0 for disconnected) | `peer_name`, `peer_id`, `user_id`, `connected` |

### Group Metrics Table

| Metric Name                              | Type      | Description                                              | Labels                                    |
| ---------------------------------------- | --------- | -------------------------------------------------------- | ----------------------------------------- |
| `netbird_groups`                         | Gauge     | Total number of NetBird groups                           | -                                         |
| `netbird_group_peers_count`              | Gauge     | Number of peers in each NetBird group                    | `group_id`, `group_name`, `issued`        |
| `netbird_group_resources_count`          | Gauge     | Number of resources in each NetBird group                | `group_id`, `group_name`, `issued`        |
| `netbird_group_info`                     | Gauge     | Information about NetBird groups (always 1)              | `group_id`, `group_name`, `issued`        |
| `netbird_group_resources_by_type`        | Gauge     | Number of resources in each group by resource type       | `group_id`, `group_name`, `resource_type` |
| `netbird_groups_scrape_errors_total`     | Counter   | Total number of errors encountered while scraping groups | `error_type`                              |
| `netbird_groups_scrape_duration_seconds` | Histogram | Time spent scraping groups from the NetBird API          | -                                         |

### User Metrics Table

| Metric Name                             | Type      | Description                                             | Labels                                                   |
| --------------------------------------- | --------- | ------------------------------------------------------- | -------------------------------------------------------- |
| `netbird_users`                         | Gauge     | Total number of NetBird users                           | -                                                        |
| `netbird_users_by_role`                 | Gauge     | Number of users by role                                 | `role`                                                   |
| `netbird_users_by_status`               | Gauge     | Number of users by status                               | `status`                                                 |
| `netbird_users_service_users`           | Gauge     | Number of service users vs regular users                | `is_service_user`                                        |
| `netbird_users_blocked`                 | Gauge     | Number of blocked vs unblocked users                    | `is_blocked`                                             |
| `netbird_users_by_issued`               | Gauge     | Number of users by issuance type                        | `issued`                                                 |
| `netbird_users_restricted`              | Gauge     | Number of users with restricted permissions             | `is_restricted`                                          |
| `netbird_user_last_login_timestamp`     | Gauge     | Last login timestamp for each user                      | `user_id`, `user_email`, `user_name`                     |
| `netbird_user_auto_groups_count`        | Gauge     | Number of auto groups assigned to each user             | `user_id`, `user_email`, `user_name`                     |
| `netbird_user_permissions`              | Gauge     | User permissions by module and action                   | `user_id`, `user_email`, `module`, `permission`, `value` |
| `netbird_users_scrape_errors_total`     | Counter   | Total number of errors encountered while scraping users | `error_type`                                             |
| `netbird_users_scrape_duration_seconds` | Histogram | Time spent scraping users from the NetBird API          | -                                                        |

### DNS Metrics Table

| Metric Name                                    | Type  | Description                                           | Labels                   |
| ---------------------------------------------- | ----- | ----------------------------------------------------- | ------------------------ |
| `netbird_dns_nameserver_groups`                | Gauge | Total number of NetBird nameserver groups             | -                        |
| `netbird_dns_nameserver_groups_enabled`        | Gauge | Number of enabled/disabled nameserver groups          | `enabled`                |
| `netbird_dns_nameserver_groups_primary`        | Gauge | Number of primary/secondary nameserver groups         | `primary`                |
| `netbird_dns_nameserver_group_domains_count`   | Gauge | Number of domains configured in each nameserver group | `group_id`, `group_name` |
| `netbird_dns_nameservers`                      | Gauge | Total number of nameservers in each group             | `group_id`, `group_name` |
| `netbird_dns_nameservers_by_type`              | Gauge | Number of nameservers by type (UDP/TCP)               | `ns_type`                |
| `netbird_dns_nameservers_by_port`              | Gauge | Number of nameservers by port                         | `port`                   |
| `netbird_dns_management_disabled_groups_count` | Gauge | Number of groups with DNS management disabled         | -                        |
### Accounts Metrics Table

| Metric Name                            | Type      | Description                                                   | Labels                                           |
| -------------------------------------- | --------- | ------------------------------------------------------------- | ------------------------------------------------ |
| `netbird_account_info`                 | Gauge     | Information about the NetBird account (always 1)              | `account_id`, `domain`, `domain_category`, `created_by` |
| `netbird_account_created_at_timestamp` | Gauge     | Unix timestamp when the account was created                   | `account_id`, `domain`                           |
| `netbird_accounts_scrape_errors_total` | Counter   | Total number of errors encountered while scraping accounts    | `error_type`                                     |
| `netbird_accounts_scrape_duration_seconds` | Histogram | Time spent scraping accounts from the NetBird API         | -                                                |

### Setup Keys Metrics Table

| Metric Name                               | Type      | Description                                                     | Labels                                                      |
| ----------------------------------------- | --------- | --------------------------------------------------------------- | ----------------------------------------------------------- |
| `netbird_setup_keys`                      | Gauge     | Total number of NetBird setup keys                              | -                                                           |
| `netbird_setup_keys_by_type`              | Gauge     | Number of NetBird setup keys by type (one-off or reusable)      | `type`                                                      |
| `netbird_setup_keys_by_state`             | Gauge     | Number of NetBird setup keys by state (valid, expired, overused, revoked) | `state`                                          |
| `netbird_setup_key_info`                  | Gauge     | Information about NetBird setup keys (always 1)                 | `key_id`, `key_name`, `type`, `state`, `valid`, `revoked`, `ephemeral` |
| `netbird_setup_key_used_times`            | Gauge     | Number of times a setup key has been used                       | `key_id`, `key_name`                                        |
| `netbird_setup_key_usage_limit`           | Gauge     | Usage limit for a setup key (0 = unlimited)                     | `key_id`, `key_name`                                        |
| `netbird_setup_key_expires_at_timestamp`  | Gauge     | Unix timestamp when the setup key expires                       | `key_id`, `key_name`                                        |
| `netbird_setup_key_last_used_timestamp`   | Gauge     | Unix timestamp when the setup key was last used                 | `key_id`, `key_name`                                        |
| `netbird_setup_keys_scrape_errors_total`  | Counter   | Total number of errors encountered while scraping setup keys    | `error_type`                                                |
| `netbird_setup_keys_scrape_duration_seconds` | Histogram | Time spent scraping setup keys from the NetBird API          | -                                                           |

### Policies Metrics Table

| Metric Name                               | Type      | Description                                                     | Labels                          |
| ----------------------------------------- | --------- | --------------------------------------------------------------- | ------------------------------- |
| `netbird_policies`                        | Gauge     | Total number of NetBird policies                                | -                               |
| `netbird_policies_by_status`              | Gauge     | Number of NetBird policies by status (enabled/disabled)         | `enabled`                       |
| `netbird_policy_info`                     | Gauge     | Information about NetBird policies (always 1)                   | `policy_id`, `policy_name`, `enabled` |
| `netbird_policy_rules_count`              | Gauge     | Number of rules in each NetBird policy                          | `policy_id`, `policy_name`      |
| `netbird_policy_posture_checks_count`     | Gauge     | Number of source posture checks in each NetBird policy          | `policy_id`, `policy_name`      |
| `netbird_policies_scrape_errors_total`    | Counter   | Total number of errors encountered while scraping policies      | `error_type`                    |
| `netbird_policies_scrape_duration_seconds` | Histogram | Time spent scraping policies from the NetBird API              | -                               |

### Routes Metrics Table

| Metric Name                              | Type      | Description                                                    | Labels                                                       |
| ---------------------------------------- | --------- | -------------------------------------------------------------- | ------------------------------------------------------------ |
| `netbird_routes`                         | Gauge     | Total number of NetBird routes                                 | -                                                            |
| `netbird_routes_by_status`               | Gauge     | Number of NetBird routes by status (enabled/disabled)          | `enabled`                                                    |
| `netbird_routes_by_network_type`         | Gauge     | Number of NetBird routes by network type                       | `network_type`                                               |
| `netbird_route_info`                     | Gauge     | Information about NetBird routes (always 1)                    | `route_id`, `network_id`, `network_type`, `enabled`, `masquerade`, `keep_route` |
| `netbird_route_metric_value`             | Gauge     | Metric value for NetBird route (lower = higher priority)       | `route_id`, `network_id`                                     |
| `netbird_route_groups_count`             | Gauge     | Number of peer groups associated with each NetBird route       | `route_id`, `network_id`                                     |
| `netbird_routes_scrape_errors_total`     | Counter   | Total number of errors encountered while scraping routes       | `error_type`                                                 |
| `netbird_routes_scrape_duration_seconds` | Histogram | Time spent scraping routes from the NetBird API                | -                                                            |

### Posture Checks Metrics Table

| Metric Name                                      | Type      | Description                                                           | Labels                     |
| ------------------------------------------------ | --------- | --------------------------------------------------------------------- | -------------------------- |
| `netbird_posture_checks`                         | Gauge     | Total number of NetBird posture checks                                | -                          |
| `netbird_posture_check_info`                     | Gauge     | Information about NetBird posture checks (always 1)                   | `check_id`, `check_name`   |
| `netbird_posture_checks_scrape_errors_total`     | Counter   | Total number of errors encountered while scraping posture checks      | `error_type`               |
| `netbird_posture_checks_scrape_duration_seconds` | Histogram | Time spent scraping posture checks from the NetBird API               | -                          |

### DNS Zones Metrics Table

| Metric Name                                   | Type      | Description                                                       | Labels                               |
| --------------------------------------------- | --------- | ----------------------------------------------------------------- | ------------------------------------ |
| `netbird_dns_zones`                           | Gauge     | Total number of NetBird DNS zones                                 | -                                    |
| `netbird_dns_zone_info`                       | Gauge     | Information about NetBird DNS zones (always 1)                    | `zone_id`, `zone_name`, `domain`, `enabled` |
| `netbird_dns_zone_records_count`              | Gauge     | Number of DNS records in each NetBird DNS zone                    | `zone_id`, `zone_name`               |
| `netbird_dns_zones_scrape_errors_total`       | Counter   | Total number of errors encountered while scraping DNS zones       | `error_type`                         |
| `netbird_dns_zones_scrape_duration_seconds`   | Histogram | Time spent scraping DNS zones from the NetBird API                | -                                    |

### Tokens Metrics Table (Opt-in)

> **Note:** This exporter is disabled by default due to high cardinality. Enable with `ENABLE_TOKENS_EXPORTER=true`

| Metric Name                                | Type      | Description                                                      | Labels                              |
| ------------------------------------------ | --------- | ---------------------------------------------------------------- | ----------------------------------- |
| `netbird_tokens`                           | Gauge     | Total number of NetBird personal access tokens across all users  | -                                   |
| `netbird_tokens_by_user`                   | Gauge     | Number of personal access tokens per user                        | `user_id`                           |
| `netbird_token_info`                       | Gauge     | Information about NetBird personal access tokens (always 1)      | `token_id`, `token_name`, `user_id` |
| `netbird_token_expires_at_timestamp`       | Gauge     | Unix timestamp when the personal access token expires            | `token_id`, `token_name`, `user_id` |
| `netbird_token_last_used_timestamp`        | Gauge     | Unix timestamp when the personal access token was last used      | `token_id`, `token_name`, `user_id` |
| `netbird_tokens_scrape_errors_total`       | Counter   | Total number of errors encountered while scraping tokens         | `error_type`                        |
| `netbird_tokens_scrape_duration_seconds`   | Histogram | Time spent scraping tokens from the NetBird API                  | -                                   |

### Events Metrics Table (Opt-in)

> **Note:** This exporter is disabled by default due to **very high cardinality**. Enable with `ENABLE_EVENTS_EXPORTER=true`. Use with caution in production.

| Metric Name                               | Type      | Description                                                     | Labels                                                                 |
| ----------------------------------------- | --------- | --------------------------------------------------------------- | ---------------------------------------------------------------------- |
| `netbird_events`                          | Gauge     | Total number of NetBird events retrieved                        | -                                                                      |
| `netbird_events_by_activity`              | Gauge     | Number of NetBird events by activity type                       | `activity_code`                                                        |
| `netbird_event_info`                      | Gauge     | Information about NetBird events (always 1)                     | `event_id`, `activity`, `activity_code`, `initiator_id`, `initiator_email`, `target_id` |
| `netbird_event_timestamp`                 | Gauge     | Unix timestamp when the event occurred                          | `event_id`, `activity_code`                                            |
| `netbird_events_scrape_errors_total`      | Counter   | Total number of errors encountered while scraping events        | `error_type`                                                           |
| `netbird_events_scrape_duration_seconds`  | Histogram | Time spent scraping events from the NetBird API                 | -                                                                      |



### Network Metrics Table

| Metric Name                                | Type      | Description                                                | Labels                                      |
| ------------------------------------------ | --------- | ---------------------------------------------------------- | ------------------------------------------- |
| `netbird_networks`                         | Gauge     | Total number of networks in your NetBird deployment        | -                                           |
| `netbird_network_routers_count`            | Gauge     | Number of routers configured in each network               | `network_id`, `network_name`                |
| `netbird_network_resources_count`          | Gauge     | Number of resources associated with each network           | `network_id`, `network_name`                |
| `netbird_network_policies_count`           | Gauge     | Number of policies applied to each network                 | `network_id`, `network_name`                |
| `netbird_network_routing_peers_count`      | Gauge     | Number of routing peers in each network                    | `network_id`, `network_name`                |
| `netbird_network_info`                     | Gauge     | Information about networks (always 1)                      | `network_id`, `network_name`, `description` |
| `netbird_networks_scrape_errors_total`     | Counter   | Total number of errors encountered while scraping networks | `error_type`                                |
| `netbird_networks_scrape_duration_seconds` | Histogram | Time spent scraping networks from the NetBird API          | -                                           |

### Exporter Metrics Table

| Metric Name                                | Type      | Description                     | Labels |
| ------------------------------------------ | --------- | ------------------------------- | ------ |
| `netbird_exporter_scrape_duration_seconds` | Histogram | Time spent scraping NetBird API | -      |
| `netbird_exporter_scrape_errors_total`     | Counter   | Total number of scrape errors   | -      |

## Configuration

The exporter is configured via environment variables:

| Variable                    | Default                  | Required | Description                                              |
| --------------------------- | ------------------------ | -------- | -------------------------------------------------------- |
| `NETBIRD_API_URL`           | `https://api.netbird.io` | No       | NetBird API base URL                                     |
| `NETBIRD_API_TOKEN`         | -                        | **Yes**  | NetBird API authentication token                         |
| `LISTEN_ADDRESS`            | `:8080`                  | No       | Address and port to listen on                            |
| `METRICS_PATH`              | `/metrics`               | No       | Path where metrics are exposed                           |
| `LOG_LEVEL`                 | `info`                   | No       | Log level (debug, info, warn, error)                     |
| `ENABLE_TOKENS_EXPORTER`    | `false`                  | No       | Enable personal access tokens metrics (high cardinality) |
| `ENABLE_EVENTS_EXPORTER`    | `false`                  | No       | Enable events/audit log metrics (very high cardinality)  |

## Getting Your NetBird API Token

1. Create a new service user with PAT with appropriate permissions. See docs: [NetBird Service Users Guide](https://docs.netbird.io/how-to/access-netbird-public-api#creating-a-service-user).
2. Copy the token and use it as `NETBIRD_API_TOKEN`

## Installation & Usage

### Option 1: Docker Compose (Recommended)

1. Clone this repository:

```bash
git clone https://github.com/matanbaruch/netbird-api-exporter
cd netbird-api-exporter
```

1. Create environment file:

```bash
cp env.example .env
# Edit .env with your NetBird API token
```

1. Start the exporter:

```bash
docker-compose up -d
```

### Option 2: Helm Chart

#### From Artifact Hub (Recommended)

Browse and install from [Artifact Hub](https://artifacthub.io/packages/helm/netbird-api-exporter/netbird-api-exporter). See our [Artifact Hub guide](docs/artifacthub.md) for more details:

```bash
# Install directly from OCI registry
helm upgrade --install netbird-api-exporter \
  oci://ghcr.io/matanbaruch/netbird-api-exporter/charts/netbird-api-exporter \
  --set netbird.apiToken=your_token_here
```

#### From GitHub Packages

Install using Helm with the chart from GitHub packages:

```bash
# Add the chart repository
helm upgrade --install netbird-api-exporter \
  oci://ghcr.io/matanbaruch/netbird-api-exporter/charts/netbird-api-exporter \
  --set netbird.apiToken=your_token_here
```

Or with a values file:

```bash
# Create values.yaml
cat <<EOF > values.yaml
netbird:
  apiToken: "your_token_here"
  apiUrl: "https://api.netbird.io"

service:
  type: ClusterIP
  port: 8080

serviceMonitor:
  enabled: true  # if using Prometheus operator
EOF

# Install the chart
helm upgrade --install netbird-api-exporter \
  oci://ghcr.io/matanbaruch/netbird-api-exporter/charts/netbird-api-exporter \
  -f values.yaml
```

### Option 3: Docker

Use the pre-built image from GitHub packages:

```bash
docker run -d \
  -p 8080:8080 \
  -e NETBIRD_API_TOKEN=your_token_here \
  --name netbird-api-exporter \
  ghcr.io/matanbaruch/netbird-api-exporter:latest
```

Or build from source:

```bash
docker build -t netbird-api-exporter .
docker run -d \
  -p 8080:8080 \
  -e NETBIRD_API_TOKEN=your_token_here \
  --name netbird-api-exporter \
  netbird-api-exporter
```

### Option 4: Go Binary

1. Install dependencies:

```bash
go mod download
```

1. Build and run:

```bash
export NETBIRD_API_TOKEN=your_token_here
go build -o netbird-api-exporter
./netbird-api-exporter
```

## Endpoints

- **`/metrics`** - Prometheus metrics endpoint
- **`/health`** - Health check endpoint (returns JSON)
- **`/`** - Information page with links

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: "netbird-api-exporter"
    static_configs:
      - targets: ["localhost:8080"]
    scrape_interval: 30s
    metrics_path: /metrics
```

## Example Queries

Here are some useful Prometheus queries:

### Peer Queries

```promql
# Total number of peers
netbird_peers

# Percentage of connected peers
(netbird_peers_connected{connected="true"} / netbird_peers) * 100

# Peers by operating system
sum by (os) (netbird_peers_by_os)

# Peers that haven't been seen in over 1 hour
(time() - netbird_peer_last_seen_timestamp) > 3600

# Number of peers requiring approval
netbird_peers_approval_required{approval_required="true"}

# Average accessible peers per peer
avg(netbird_peer_accessible_peers_count)

# Connection status of specific peer by name
netbird_peer_connection_status_by_name{peer_name="aura-netbird-us-east-1-eks-infra-0"}

# All disconnected peers
netbird_peer_connection_status_by_name{connected="false"}

# All connected peers
netbird_peer_connection_status_by_name{connected="true"}

# Count of connected vs disconnected peers by name
sum(netbird_peer_connection_status_by_name) by (connected)
```

### Group Queries

```promql
# Total number of groups
netbird_groups

# Groups with the most peers
topk(5, netbird_group_peers_count)

# Groups with the most resources
topk(5, netbird_group_resources_count)

# Average peers per group
avg(netbird_group_peers_count)

# Groups by issued method (API vs manual)
count by (issued) (netbird_group_info)

# Resource distribution by type across all groups
sum by (resource_type) (netbird_group_resources_by_type)

# Groups with no peers
netbird_group_peers_count == 0

# Groups with no resources
netbird_group_resources_count == 0

# Groups scrape error rate
rate(netbird_groups_scrape_errors_total[5m])
```

### User Queries

```promql
# Total number of users
netbird_users

# Users by role
sum by (role) (netbird_users_by_role)

# Users by status
sum by (status) (netbird_users_by_status)

# Service users vs regular users
netbird_users_service_users

# Blocked users
netbird_users_blocked

# Users by issuance type
sum by (issued) (netbird_users_by_issued)

# Users with restricted permissions
netbird_users_restricted

# Last login timestamp for each user
netbird_user_last_login_timestamp

# Auto groups assigned to each user
netbird_user_auto_groups_count

# User permissions by module and action
sum by (module, permission) (netbird_user_permissions)
```

### DNS Queries

```promql
# Total number of nameserver groups
netbird_dns_nameserver_groups

# Enabled vs disabled nameserver groups
netbird_dns_nameserver_groups_enabled

# Primary vs secondary nameserver groups
netbird_dns_nameserver_groups_primary

# Nameserver groups with the most domains
topk(5, netbird_dns_nameserver_group_domains_count)

# Nameserver groups with the most nameservers
topk(5, netbird_dns_nameservers)

# Nameserver distribution by type
sum by (ns_type) (netbird_dns_nameservers_by_type)

# Nameserver distribution by port
sum by (port) (netbird_dns_nameservers_by_port)

# Groups with DNS management disabled
netbird_dns_management_disabled_groups_count

# Average domains per nameserver group
avg(netbird_dns_nameserver_group_domains_count)

# Nameserver groups with no domains configured
netbird_dns_nameserver_group_domains_count == 0

# Total nameservers across all groups
sum(netbird_dns_nameservers)
```

### Network Queries

```promql
# Total number of networks
netbird_networks

# Networks with the most routers
topk(5, netbird_network_routers_count)

# Networks with the most resources
topk(5, netbird_network_resources_count)

# Networks with the most policies
topk(5, netbird_network_policies_count)

# Networks with the most routing peers
topk(5, netbird_network_routing_peers_count)

# Average routers per network
avg(netbird_network_routers_count)

# Average resources per network
avg(netbird_network_resources_count)

# Networks with no routers
netbird_network_routers_count == 0

# Networks with no resources
netbird_network_resources_count == 0

# Networks with no policies
netbird_network_policies_count == 0

# Total routers across all networks
sum(netbird_network_routers_count)

# Total resources across all networks
sum(netbird_network_resources_count)

# Networks scrape error rate
rate(netbird_networks_scrape_errors_total[5m])
```

## Grafana Dashboard

A comprehensive pre-built Grafana dashboard is available that provides visualizations for all NetBird API Exporter metrics.

### Quick Start

1. **Download the dashboard**: Get [`grafana-dashboard.json`](grafana-dashboard.json) from this repository
2. **Import in Grafana**: Go to Dashboards → Import and upload the JSON file
3. **Configure data source**: Ensure your Prometheus data source is selected

### Dashboard Features

The dashboard includes organized sections for:

- **Overview**: Key metrics summary (total peers, users, groups, networks)
- **Peers**: Connection status, OS distribution, geographic breakdown
- **Users**: Role distribution, status overview, service vs regular users
- **Groups**: Peer and resource counts per group
- **DNS**: Nameserver configurations and status
- **Networks**: Network information and resource distribution
- **Performance**: API response times and error rates

### Documentation

For detailed installation instructions, customization options, and troubleshooting, see the [Grafana Dashboard Documentation](docs/grafana-dashboard.md).

### Manual Dashboard Creation

If you prefer to create custom panels, here are some example configurations:

## Troubleshooting

### Common Issues

1. **Authentication errors**: Verify your `NETBIRD_API_TOKEN` is correct and has appropriate permissions
2. **Connection errors**: Check if the NetBird API URL is accessible from your network
3. **Missing metrics**: Ensure your NetBird account has peers registered

### Logs

Check logs for debugging:

```bash
# Docker Compose
docker-compose logs netbird-api-exporter

# Docker
docker logs netbird-api-exporter

# Binary
# Logs are output to stdout
```

### Enable Debug Logging

Set `LOG_LEVEL=debug` for more verbose output.

## Security Considerations

- Store your NetBird API token securely (use Docker secrets, Kubernetes secrets, etc.)
- Consider running the exporter in a private network
- Implement proper firewall rules to restrict access to the metrics endpoint
- Regularly rotate your API tokens

### Artifact Verification

All releases include signed build provenance attestations for enhanced supply chain security. You can verify the authenticity of our artifacts using the GitHub CLI:

```bash
# Verify Docker image attestation
gh attestation verify oci://ghcr.io/matanbaruch/netbird-api-exporter:latest --owner matanbaruch

# Download and verify binary attestations
gh run download --repo matanbaruch/netbird-api-exporter --name netbird-api-exporter-binaries-[VERSION]
gh attestation verify netbird-api-exporter-linux-amd64 --owner matanbaruch
```

For complete security documentation, see [SECURITY.md](SECURITY.md).

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed list of changes, new features, and bug fixes in each release.

## Development

### Prerequisites

- Go 1.23 or later
- golangci-lint (for linting)

### Building from Source

```bash
go mod download
go build -o netbird-api-exporter
```

### Code Quality

This project uses several tools to maintain code quality:

#### Pre-commit Hooks (Recommended)

Set up pre-commit hooks to automatically run formatting, linting, and tests before each commit:

```bash
make setup-precommit
```

This provides two options:

- **Simple Git Hook**: Basic bash script (no external dependencies)
- **Pre-commit Framework**: Advanced hook management with additional checks

For quick setup with the simple git hook:

```bash
make install-hooks
```

See the [Pre-commit Hooks Guide](docs/getting-started/pre-commit-hooks.md) for detailed setup instructions and configuration options.

#### Linting

Run linting checks:

```bash
make lint
```

This runs:

- `golangci-lint` - Comprehensive Go linting
- `go vet` - Go's built-in static analysis
- `gofmt` - Code formatting check

#### Formatting

Format code:

```bash
make fmt
```

#### Running All Checks

Run tests and linting together:

```bash
make check
```

#### Available Make Targets

```bash
make help
```

Shows all available targets including:

- `build` - Build the binary
- `test` - Run tests
- `lint` - Run linting checks
- `fmt` - Format code
- `check` - Run all checks (tests + linting)

### Continuous Integration

The project includes GitHub Actions workflows that automatically:

- Run linting checks on all pull requests
- Verify code formatting
- Run tests (using self-hosted NetBird instance)
- Check for security issues

### Running Tests

```bash
go test ./...
```

## Stargazers over time

[![Stargazers over time](https://starchart.cc/matanbaruch/netbird-api-exporter.svg?variant=adaptive)](https://starchart.cc/matanbaruch/netbird-api-exporter)
