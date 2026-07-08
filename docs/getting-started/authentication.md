---
layout: default
title: Authentication
parent: Getting Started
nav_order: 1
---

# Authentication
{: .no_toc }

Learn how to obtain and configure your NetBird API token for the exporter.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Getting Your NetBird API Token

The NetBird API Exporter requires a valid API token to access your NetBird deployment's data. Follow these steps to create and configure your token.

### Step 1: Access NetBird Dashboard

1. Open your web browser and navigate to your NetBird dashboard
2. Log in with your NetBird account credentials
3. Ensure you have administrative privileges (required for API key creation)

### Step 2: Navigate to API Keys

1. In the NetBird dashboard, click on **Settings** in the left sidebar
2. Select **API Keys** from the settings menu
3. You'll see a list of existing API keys (if any)

### Step 3: Create New API Key

1. Click the **"Add API Key"** or **"Create New Key"** button
2. Fill in the API key details:
   - **Name**: Give it a descriptive name (e.g., "prometheus-exporter")
   - **Description**: Optional description for future reference
   - **Expiration**: Set an appropriate expiration date (recommended: 1 year)

### Step 4: Set Permissions

Configure the appropriate permissions for the API key. The exporter needs **read-only** access to the following resources:

#### Required Permissions
- ✅ **Peers**: Read access to view peer information
- ✅ **Groups**: Read access to view group information
- ✅ **Users**: Read access to view user information
- ✅ **DNS**: Read access to view DNS configuration
- ✅ **Networks**: Read access to view network information
- ✅ **Setup Keys**: Read access to view setup key information
- ✅ **Policies**: Read access to view policy information
- ✅ **Routes**: Read access to view route information

#### Permission Settings
```
Peers: Read Only ✓
Groups: Read Only ✓
Users: Read Only ✓
DNS: Read Only ✓
Networks: Read Only ✓
Setup Keys: Read Only ✓
Policies: Read Only ✓
Routes: Read Only ✓
Settings: No Access (not needed)
```

{: .warning }
> **Security Best Practice**: Only grant the minimum permissions required. The exporter only needs read access and should never be given write permissions.

### Step 5: Generate and Copy Token

1. Click **"Create"** or **"Generate"** to create the API key
2. **Important**: Copy the generated token immediately and store it securely
3. The token will look similar to: `nb_api_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

{: .important }
> **Token Security**: This is the only time you'll see the full token. Store it securely and never share it publicly.

## Configuring the Token

Once you have your API token, you need to configure it for the exporter. The method depends on your deployment approach:

### Environment Variable (All Methods)

Set the token as an environment variable:

```bash
export NETBIRD_API_TOKEN="nb_api_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

### Docker Compose

Add to your `.env` file:

```bash
NETBIRD_API_TOKEN=nb_api_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### Docker

Pass as environment variable:

```bash
docker run -e NETBIRD_API_TOKEN="nb_api_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" ...
```

### Kubernetes/Helm

Create a secret:

```bash
kubectl create secret generic netbird-api-token \
  --from-literal=token="nb_api_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

### systemd

Add to environment file (`/etc/netbird-api-exporter/config`):

```bash
NETBIRD_API_TOKEN=nb_api_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Testing Your Token

Verify your token works correctly by testing it directly with the NetBird API:

```bash
curl -H "Authorization: Bearer nb_api_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
     https://api.netbird.io/api/peers
```

Expected response (truncated):
```json
{
  "peers": [
    {
      "id": "peer-123",
      "name": "my-peer",
      "ip": "100.64.0.1",
      "connected": true,
      ...
    }
  ]
}
```

## Token Management

### Token Rotation

For security, regularly rotate your API tokens:

1. Create a new API key following the steps above
2. Update your exporter configuration with the new token
3. Test that the exporter works with the new token
4. Delete the old API key from the NetBird dashboard

### Multiple Tokens

You can create multiple API keys for different purposes:

- **Production Exporter**: Long-lived token for production monitoring
- **Development Exporter**: Short-lived token for testing
- **Backup Token**: Emergency backup token

### Token Monitoring

Monitor your API token usage:

1. Check token expiration dates regularly
2. Review API key access logs in NetBird dashboard
3. Set up alerts for token expiration
4. Monitor exporter logs for authentication errors

## Troubleshooting Authentication

### Common Issues

#### 1. Invalid Token Error
```
Error: 401 Unauthorized - Invalid API token
```

**Solutions**:
- Verify the token is copied correctly (no extra spaces/characters)
- Check token hasn't expired
- Ensure token has proper permissions

#### 2. Insufficient Permissions
```
Error: 403 Forbidden - Insufficient permissions
```

**Solutions**:
- Review API key permissions in NetBird dashboard
- Ensure all required read permissions are granted
- Recreate token with correct permissions

#### 3. Token Expired
```
Error: 401 Unauthorized - Token has expired
```

**Solutions**:
- Create a new API token
- Update exporter configuration
- Set longer expiration for future tokens

### Debugging Steps

1. **Verify token format**: Should start with `nb_api_`
2. **Test token manually**: Use curl to test API access
3. **Check exporter logs**: Look for authentication error messages
4. **Verify permissions**: Ensure token has all required read permissions

## Security Considerations

### Token Storage

- **Never commit tokens to version control**
- **Use secure secret management** (Kubernetes secrets, Docker secrets, etc.)
- **Restrict access** to token storage locations
- **Encrypt tokens at rest** when possible

### Network Security

- **Use HTTPS only** for API communication
- **Implement firewall rules** to restrict exporter network access
- **Consider VPN/private networks** for additional security

### Monitoring

- **Log authentication events** for security auditing
- **Monitor for unusual API usage** patterns
- **Set up alerts** for authentication failures
- **Regular security reviews** of API token usage

## Next Steps

Once you have your API token configured:

1. **[Choose your installation method](../installation/)** 
2. **[Configure the exporter](../installation/docker-compose)** with your token
3. **[Verify the installation](../getting-started#quick-verification)** is working correctly 
