# Production Deployment Security Guide

This guide provides security recommendations and best practices for deploying NetBird API Exporter in production environments.

## Table of Contents

1. [Quick Security Checklist](#quick-security-checklist)
2. [Authentication](#authentication)
3. [TLS/HTTPS Configuration](#tlshttps-configuration)
4. [Network Security](#network-security)
5. [Container Security](#container-security)
6. [Secret Management](#secret-management)
7. [Monitoring and Alerting](#monitoring-and-alerting)

---

## Quick Security Checklist

Before deploying to production, ensure:

- [ ] Metrics endpoint is protected (authentication or network restriction)
- [ ] TLS/HTTPS is enabled (via reverse proxy or ingress)
- [ ] API token is stored in a secrets management system
- [ ] Running as non-root user (UID 65534)
- [ ] Network policies restrict access to known Prometheus instances
- [ ] Log level is set to `info` or `warn` (not `debug`)
- [ ] Container image has been scanned for vulnerabilities
- [ ] Resource limits are configured
- [ ] Monitoring and alerting are set up

---

## Authentication

### Problem: Unauthenticated Metrics Endpoint

By default, the `/metrics` endpoint has NO authentication. Any client that can reach the exporter can access all metrics, which may contain sensitive infrastructure information.

### Solution 1: Network-Level Access Control (Recommended)

Use Kubernetes NetworkPolicies to restrict access:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: netbird-exporter-access
  namespace: monitoring
spec:
  podSelector:
    matchLabels:
      app: netbird-api-exporter
  policyTypes:
  - Ingress
  ingress:
  # Allow from Prometheus only
  - from:
    - podSelector:
        matchLabels:
          app: prometheus
    ports:
    - protocol: TCP
      port: 8080
  # Allow from within same namespace for debugging (optional)
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 8080
```

### Solution 2: Reverse Proxy with Basic Auth

Deploy behind nginx with authentication:

```nginx
# /etc/nginx/conf.d/exporter.conf
server {
    listen 443 ssl http2;
    server_name metrics.example.com;

    ssl_certificate /etc/nginx/certs/cert.pem;
    ssl_certificate_key /etc/nginx/certs/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location /metrics {
        auth_basic "Prometheus Metrics";
        auth_basic_user_file /etc/nginx/.htpasswd;

        proxy_pass http://netbird-exporter:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        # Health check doesn't need auth
        proxy_pass http://netbird-exporter:8080;
    }
}
```

Create htpasswd file:

```bash
htpasswd -c /etc/nginx/.htpasswd prometheus
```

Configure Prometheus to use authentication:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'netbird-exporter'
    scheme: https
    basic_auth:
      username: prometheus
      password_file: /etc/prometheus/exporter-password
    static_configs:
      - targets: ['metrics.example.com']
```

### Solution 3: OAuth2 Proxy

Use oauth2-proxy for SSO authentication:

```yaml
# oauth2-proxy deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oauth2-proxy
spec:
  template:
    spec:
      containers:
      - name: oauth2-proxy
        image: quay.io/oauth2-proxy/oauth2-proxy:latest
        args:
        - --provider=google
        - --email-domain=example.com
        - --upstream=http://netbird-exporter:8080
        - --http-address=0.0.0.0:4180
        - --cookie-secret=CHANGEME
        env:
        - name: OAUTH2_PROXY_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: oauth2-proxy
              key: client-id
        - name: OAUTH2_PROXY_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: oauth2-proxy
              key: client-secret
```

---

## TLS/HTTPS Configuration

### Kubernetes Ingress with cert-manager

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: netbird-exporter
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - metrics.example.com
    secretName: netbird-exporter-tls
  rules:
  - host: metrics.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: netbird-api-exporter
            port:
              number: 8080
```

### Docker Compose with Traefik

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v2.10
    command:
      - "--providers.docker=true"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.email=admin@example.com"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
    ports:
      - "443:443"
      - "80:80"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./letsencrypt:/letsencrypt

  netbird-exporter:
    image: netbird-api-exporter:latest
    environment:
      NETBIRD_API_TOKEN: ${NETBIRD_API_TOKEN}
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.exporter.rule=Host(`metrics.example.com`)"
      - "traefik.http.routers.exporter.entrypoints=websecure"
      - "traefik.http.routers.exporter.tls.certresolver=letsencrypt"
```

---

## Network Security

### Kubernetes Network Policies

Complete network isolation example:

```yaml
---
# Deny all ingress by default
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-ingress
  namespace: monitoring
spec:
  podSelector: {}
  policyTypes:
  - Ingress

---
# Allow NetBird API access (egress)
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: netbird-exporter-egress
  namespace: monitoring
spec:
  podSelector:
    matchLabels:
      app: netbird-api-exporter
  policyTypes:
  - Egress
  egress:
  # DNS
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    - podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - protocol: UDP
      port: 53
  # NetBird API
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443
  # Allow internal metrics scraping
  - to:
    - podSelector:
        matchLabels:
          app: prometheus
    ports:
    - protocol: TCP
      port: 8080
```

### Firewall Rules (iptables)

```bash
#!/bin/bash
# Allow only Prometheus to access metrics

PROMETHEUS_IP="10.0.1.100"
EXPORTER_PORT=8080

# Drop all existing rules for this port
iptables -D INPUT -p tcp --dport $EXPORTER_PORT -j ACCEPT 2>/dev/null || true
iptables -D INPUT -p tcp --dport $EXPORTER_PORT -j DROP 2>/dev/null || true

# Allow from Prometheus
iptables -A INPUT -p tcp --dport $EXPORTER_PORT -s $PROMETHEUS_IP -j ACCEPT

# Allow from localhost (for health checks)
iptables -A INPUT -p tcp --dport $EXPORTER_PORT -s 127.0.0.1 -j ACCEPT

# Drop everything else
iptables -A INPUT -p tcp --dport $EXPORTER_PORT -j DROP

# Save rules
iptables-save > /etc/iptables/rules.v4
```

---

## Container Security

### Kubernetes Security Context

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: netbird-api-exporter
spec:
  template:
    spec:
      # Pod-level security context
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        runAsGroup: 65534
        fsGroup: 65534
        seccompProfile:
          type: RuntimeDefault

      containers:
      - name: exporter
        image: netbird-api-exporter:latest

        # Container-level security context
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 65534
          runAsGroup: 65534
          capabilities:
            drop:
            - ALL

        # Resource limits
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"

        # Volume mounts for read-only root filesystem
        volumeMounts:
        - name: tmp
          mountPath: /tmp
        - name: app-temp
          mountPath: /app/tmp

      volumes:
      - name: tmp
        emptyDir:
          sizeLimit: 100Mi
      - name: app-temp
        emptyDir:
          sizeLimit: 10Mi
```

### Docker Security Options

```bash
docker run -d \
  --name netbird-exporter \
  --security-opt=no-new-privileges:true \
  --cap-drop=ALL \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=100m \
  --memory=256m \
  --memory-swap=256m \
  --cpu-shares=512 \
  --pids-limit=100 \
  -e NETBIRD_API_TOKEN="${NETBIRD_API_TOKEN}" \
  -p 127.0.0.1:8080:8080 \
  netbird-api-exporter:latest
```

### Image Scanning Workflow

```yaml
# .github/workflows/security-scan.yml
name: Security Scan

on:
  push:
  schedule:
    - cron: '0 0 * * 0'  # Weekly

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build image
        run: docker build -t netbird-exporter:test .

      - name: Run Trivy scan
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: netbird-exporter:test
          format: sarif
          output: trivy-results.sarif

      - name: Upload results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: trivy-results.sarif
```

---

## Secret Management

### Kubernetes with External Secrets Operator

```yaml
# Install External Secrets Operator
# https://external-secrets.io/

---
# SecretStore pointing to Vault
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
  namespace: monitoring
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "netbird-exporter"

---
# ExternalSecret syncing from Vault
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: netbird-api-token
  namespace: monitoring
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: netbird-api-token
  data:
  - secretKey: token
    remoteRef:
      key: netbird-exporter/api-token
      property: token

---
# Deployment using the secret
apiVersion: apps/v1
kind: Deployment
metadata:
  name: netbird-api-exporter
spec:
  template:
    spec:
      containers:
      - name: exporter
        image: netbird-api-exporter:latest
        env:
        - name: NETBIRD_API_TOKEN
          valueFrom:
            secretKeyRef:
              name: netbird-api-token
              key: token
```

### HashiCorp Vault Integration

```bash
# Store token in Vault
vault kv put secret/netbird-exporter/api-token \
  token="nbp_xxxxxxxxxxxxxxxxxxxxx"

# Create policy
vault policy write netbird-exporter - <<EOF
path "secret/data/netbird-exporter/api-token" {
  capabilities = ["read"]
}
EOF

# Create Kubernetes auth role
vault write auth/kubernetes/role/netbird-exporter \
  bound_service_account_names=netbird-exporter \
  bound_service_account_namespaces=monitoring \
  policies=netbird-exporter \
  ttl=24h
```

### AWS Secrets Manager

```yaml
# Install AWS Secrets Manager CSI Driver
# https://github.com/aws/secrets-store-csi-driver-provider-aws

---
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: netbird-api-token
  namespace: monitoring
spec:
  provider: aws
  parameters:
    objects: |
      - objectName: "netbird-exporter-token"
        objectType: "secretsmanager"
  secretObjects:
  - secretName: netbird-api-token
    type: Opaque
    data:
    - objectName: netbird-exporter-token
      key: token

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: netbird-api-exporter
spec:
  template:
    spec:
      serviceAccountName: netbird-exporter
      containers:
      - name: exporter
        image: netbird-api-exporter:latest
        env:
        - name: NETBIRD_API_TOKEN
          valueFrom:
            secretKeyRef:
              name: netbird-api-token
              key: token
        volumeMounts:
        - name: secrets-store
          mountPath: "/mnt/secrets-store"
          readOnly: true
      volumes:
      - name: secrets-store
        csi:
          driver: secrets-store.csi.k8s.io
          readOnly: true
          volumeAttributes:
            secretProviderClass: netbird-api-token
```

---

## Monitoring and Alerting

### Security Monitoring Rules

```yaml
# prometheus-alerts.yml
groups:
  - name: netbird_exporter_security
    interval: 1m
    rules:
      # High error rate may indicate attack
      - alert: NetBirdExporterHighErrorRate
        expr: |
          rate(netbird_exporter_scrape_errors_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
          component: netbird-exporter
        annotations:
          summary: "High error rate in NetBird API exporter"
          description: "Error rate is {{ $value | humanizePercentage }} over 5 minutes"

      # Unusual scraping activity
      - alert: NetBirdExporterUnusualActivity
        expr: |
          rate(netbird_exporter_scrape_duration_seconds_count[5m]) > 2
        for: 10m
        labels:
          severity: warning
          component: netbird-exporter
        annotations:
          summary: "Unusual scraping frequency detected"
          description: "Scrape rate is {{ $value }} per second"

      # Exporter down
      - alert: NetBirdExporterDown
        expr: up{job="netbird-exporter"} == 0
        for: 5m
        labels:
          severity: critical
          component: netbird-exporter
        annotations:
          summary: "NetBird API exporter is down"
          description: "Exporter has been down for more than 5 minutes"

      # Slow response times (potential DoS)
      - alert: NetBirdExporterSlowScrapes
        expr: |
          rate(netbird_exporter_scrape_duration_seconds_sum[5m])
          /
          rate(netbird_exporter_scrape_duration_seconds_count[5m])
          > 30
        for: 10m
        labels:
          severity: warning
          component: netbird-exporter
        annotations:
          summary: "NetBird API exporter scrapes are slow"
          description: "Average scrape duration is {{ $value }}s"
```

### Log Monitoring

```yaml
# promtail-config.yml (for Loki)
scrape_configs:
  - job_name: netbird-exporter
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - monitoring
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: netbird-api-exporter
    pipeline_stages:
      - json:
          expressions:
            level: level
            msg: msg
            error: error

      # Alert on error patterns
      - match:
          selector: '{level="error"}'
          stages:
          - metrics:
              error_total:
                type: Counter
                description: "Total number of errors"
                source: level
                config:
                  action: inc
```

### Security Dashboard (Grafana)

```json
{
  "dashboard": {
    "title": "NetBird Exporter Security",
    "panels": [
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(netbird_exporter_scrape_errors_total[5m])"
          }
        ]
      },
      {
        "title": "Scrape Frequency",
        "targets": [
          {
            "expr": "rate(netbird_exporter_scrape_duration_seconds_count[5m])"
          }
        ]
      },
      {
        "title": "Response Time P95",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(netbird_exporter_scrape_duration_seconds_bucket[5m]))"
          }
        ]
      }
    ]
  }
}
```

---

## Additional Security Measures

### Rate Limiting

Implement at reverse proxy level:

```nginx
# Rate limiting configuration
limit_req_zone $binary_remote_addr zone=metrics:10m rate=10r/s;

server {
    location /metrics {
        limit_req zone=metrics burst=20 nodelay;
        limit_req_status 429;

        # Additional security headers
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-Frame-Options "DENY" always;
        add_header X-XSS-Protection "1; mode=block" always;

        proxy_pass http://netbird-exporter:8080;
    }
}
```

### Regular Security Maintenance

Create a security maintenance schedule:

```markdown
## Monthly Tasks
- [ ] Review and update access controls
- [ ] Scan container images for vulnerabilities
- [ ] Update dependencies and base images
- [ ] Review security logs and alerts
- [ ] Test incident response procedures

## Quarterly Tasks
- [ ] Rotate API tokens
- [ ] Audit network policies
- [ ] Review and update security documentation
- [ ] Conduct security assessment
- [ ] Update disaster recovery plans

## Annual Tasks
- [ ] Comprehensive security audit
- [ ] Penetration testing
- [ ] Review compliance requirements
- [ ] Update security training
```

---

## Incident Response

### Security Incident Checklist

1. **Detect**
   - Monitor alerts for unusual activity
   - Review logs for suspicious patterns
   - Check for unauthorized access

2. **Contain**
   - Rotate API tokens immediately
   - Block suspicious IP addresses
   - Isolate affected systems

3. **Investigate**
   - Collect logs and metrics
   - Identify attack vector
   - Assess scope of compromise

4. **Remediate**
   - Apply security patches
   - Update configurations
   - Strengthen access controls

5. **Recover**
   - Restore normal operations
   - Verify system integrity
   - Monitor for recurrence

6. **Document**
   - Record incident details
   - Document lessons learned
   - Update security procedures

---

## Support and Resources

- **Security Issues**: Report privately to repository maintainers
- **Documentation**: See [SECURITY.md](../SECURITY.md)
- **Best Practices**: [Prometheus Security](https://prometheus.io/docs/operating/security/)
- **Kubernetes Security**: [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
