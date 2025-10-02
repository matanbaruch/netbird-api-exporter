---
layout: default
title: NetBird API Exporter
nav_order: 1
description: "A Prometheus exporter for NetBird API that provides comprehensive metrics about your NetBird network"
permalink: /
---

# NetBird API Exporter
{: .fs-9 }

A Prometheus exporter for NetBird API that provides comprehensive metrics about your NetBird network peers, groups, users, networks, and DNS configuration.
{: .fs-6 .fw-300 }

[Get started now](#quick-start){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View it on GitHub](https://github.com/matanbaruch/netbird-api-exporter){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## What is NetBird API Exporter?

The NetBird API Exporter is a lightweight Prometheus exporter that fetches metrics from the [NetBird REST API](https://docs.netbird.io/ipa/resources/peers) and exposes them in Prometheus format. It provides detailed insights into:

- **Peer Metrics**: Connection status, operating systems, geographic distribution, and accessibility
- **Group Metrics**: Group sizes, resource distribution, and management statistics
- **User Metrics**: User roles, statuses, permissions, and activity
- **DNS Metrics**: Nameserver groups, domains, and DNS configuration
- **Network Metrics**: Network topology, routers, resources, and policies

## Features

✅ **Comprehensive Metrics**: 40+ different metrics covering all aspects of your NetBird deployment  
✅ **Multiple Deployment Options**: Docker, Docker Compose, Helm, systemd, or native binary  
✅ **Prometheus Integration**: Native Prometheus metrics format with proper labels  
✅ **High Performance**: Efficient API calls with error handling and recovery  
✅ **Security Focused**: Minimal privileges and secure defaults  
✅ **Easy Configuration**: Simple environment variable configuration  
✅ **Health Monitoring**: Built-in health checks and self-monitoring metrics  

## Quick Start

### Prerequisites

- NetBird API token ([how to get one](getting-started/authentication))
- Prometheus server or compatible metrics collection system

### 1. Get Your API Token

1. Log into your NetBird dashboard
2. Go to **Settings** → **API Keys**
3. Create a new API key with appropriate permissions
4. Copy the token for configuration

### 2. Choose Your Deployment Method

Pick the deployment method that works best for your environment:

| Method | Best For | Complexity |
|--------|----------|------------|
| [Docker Compose](installation/docker-compose) | Development & Testing | ⭐ Easy |
| [Docker](installation/docker) | Container Environments | ⭐⭐ Medium |
| [Helm](installation/helm) | Kubernetes Clusters | ⭐⭐ Medium |
| [systemd](installation/systemd) | Linux Servers | ⭐⭐⭐ Advanced |
| [Binary](installation/binary) | Custom Setups | ⭐⭐⭐ Advanced |

### 3. Configure Prometheus

Add the exporter to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'netbird-api-exporter'
    static_configs:
      - targets: ['localhost:8080']  # Update with your exporter address
    scrape_interval: 30s
    metrics_path: /metrics
```

### 4. Start Monitoring

Once running, you can:
- View metrics at `http://localhost:8080/metrics`
- Check health at `http://localhost:8080/health`
- Create Grafana dashboards with the [example queries](usage/prometheus-queries)

---

## Architecture

The exporter is built with a modular architecture that makes it easy to extend and maintain:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   NetBird API   │◄───│  API Exporters   │◄───│   Prometheus    │
│                 │    │                  │    │                 │
│ • Peers         │    │ • Peers          │    │ • Scraping      │
│ • Groups        │    │ • Groups         │    │ • Storage       │
│ • Users         │    │ • Users          │    │ • Querying      │
│ • DNS           │    │ • DNS            │    │ • Alerting      │
│ • Networks      │    │ • Networks       │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

Learn more about the [architecture](technical/architecture) and [available metrics](reference/metrics).

---

## Need Help?

- 📖 **Documentation**: Browse the complete documentation in the sidebar
- 🐛 **Issues**: Report bugs on [GitHub Issues](https://github.com/matanbaruch/netbird-api-exporter/issues)
- 💡 **Discussions**: Ask questions in [GitHub Discussions](https://github.com/matanbaruch/netbird-api-exporter/discussions)
- 📧 **Contact**: Reach out to the maintainers

---

## Contributing

We welcome contributions! See our [Contributing Guide](contributing) for details on:
- Code of Conduct
- Development setup
- Submitting pull requests
- Reporting issues

## License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/matanbaruch/netbird-api-exporter/blob/main/LICENSE) file for details. 

