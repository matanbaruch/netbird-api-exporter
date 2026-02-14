#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

COMPOSE_FILE="tests/docker-compose.netbird.yml"
CONTAINER_NAME="netbird-server"
MAX_WAIT=90

print_status() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Resolve the project root (where this script lives is scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

print_status "Starting self-hosted NetBird server for testing..."

# Start the container
docker compose -f "$COMPOSE_FILE" up -d

# Wait for the server to be healthy
print_status "Waiting for NetBird server to be ready (max ${MAX_WAIT}s)..."
elapsed=0
while [ $elapsed -lt $MAX_WAIT ]; do
    if docker exec "$CONTAINER_NAME" wget -q --spider http://localhost:80/oauth2/.well-known/openid-configuration 2>/dev/null; then
        print_success "NetBird server is ready! (${elapsed}s)"
        break
    fi
    sleep 3
    elapsed=$((elapsed + 3))
done

if [ $elapsed -ge $MAX_WAIT ]; then
    print_error "NetBird server failed to start within ${MAX_WAIT}s"
    docker compose -f "$COMPOSE_FILE" logs
    exit 1
fi

# Create an API token
print_status "Creating API token..."
TOKEN_OUTPUT=$(docker exec "$CONTAINER_NAME" /go/bin/netbird-server token create --name "ci-test" --config /etc/netbird/config.yaml 2>&1)
TOKEN=$(echo "$TOKEN_OUTPUT" | grep "^Token:" | awk '{print $2}')

if [ -z "$TOKEN" ]; then
    print_error "Failed to create API token"
    echo "$TOKEN_OUTPUT"
    exit 1
fi

print_success "API token created successfully"

# Export environment variables
export NETBIRD_API_URL="http://localhost:8081"
export NETBIRD_API_TOKEN="$TOKEN"

# Output for CI (GitHub Actions)
if [ -n "$GITHUB_ENV" ]; then
    echo "NETBIRD_API_URL=http://localhost:8081" >> "$GITHUB_ENV"
    echo "NETBIRD_API_TOKEN=$TOKEN" >> "$GITHUB_ENV"
fi

# Output for sourcing in shell
echo ""
echo "# Source these environment variables to run integration tests:"
echo "export NETBIRD_API_URL=http://localhost:8081"
echo "export NETBIRD_API_TOKEN=$TOKEN"
