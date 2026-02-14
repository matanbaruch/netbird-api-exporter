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
NETBIRD_URL="http://localhost:8081"
ADMIN_EMAIL="admin@test.local"
ADMIN_PASSWORD="T3stP@ssw0rd!"
ADMIN_NAME="CI Admin"

print_status() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Resolve the project root (where this script lives is scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

print_status "Starting self-hosted NetBird server for testing..."

# Start the container
docker compose -f "$COMPOSE_FILE" up -d

# Wait for the server to be healthy (poll from host)
print_status "Waiting for NetBird server to be ready (max ${MAX_WAIT}s)..."
elapsed=0
while [ $elapsed -lt $MAX_WAIT ]; do
    if curl -sf "${NETBIRD_URL}/oauth2/.well-known/openid-configuration" >/dev/null 2>&1; then
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

# Step 1: Create admin user via setup endpoint
print_status "Creating admin user via /api/setup..."
SETUP_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${NETBIRD_URL}/api/setup" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\",\"name\":\"${ADMIN_NAME}\"}")

SETUP_HTTP_CODE=$(echo "$SETUP_RESPONSE" | tail -1)
SETUP_BODY=$(echo "$SETUP_RESPONSE" | sed '$d')

if [ "$SETUP_HTTP_CODE" -ge 200 ] && [ "$SETUP_HTTP_CODE" -lt 300 ]; then
    USER_ID=$(echo "$SETUP_BODY" | jq -r '.user_id // .userId // empty')
    print_success "Admin user created (HTTP ${SETUP_HTTP_CODE})"
else
    print_warning "Setup returned HTTP ${SETUP_HTTP_CODE}: ${SETUP_BODY}"
    print_status "Setup may have already been completed, continuing..."
fi

# Step 2: Get JWT via OAuth2 Authorization Code flow with PKCE
# Dex doesn't support ROPC, so we automate the auth code flow with curl
print_status "Authenticating via OAuth2 authorization code flow..."

COOKIE_JAR=$(mktemp)

# Generate PKCE code verifier and challenge
CODE_VERIFIER=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 43)
CODE_CHALLENGE=$(echo -n "$CODE_VERIFIER" | openssl dgst -sha256 -binary | openssl base64 -A | tr '+/' '-_' | tr -d '=')

# Step 2a: Start auth flow - get login page HTML and URL in a single request
AUTH_URL="${NETBIRD_URL}/oauth2/auth?client_id=netbird-cli&response_type=code&redirect_uri=http://localhost:53000/&scope=openid+profile+email&code_challenge=${CODE_CHALLENGE}&code_challenge_method=S256"

DELIM="__CURL_URL__"
LOGIN_COMBINED=$(curl -s -c "$COOKIE_JAR" -b "$COOKIE_JAR" -L -w "${DELIM}%{url_effective}" "$AUTH_URL")
LOGIN_PAGE_URL="${LOGIN_COMBINED##*${DELIM}}"
LOGIN_HTML="${LOGIN_COMBINED%%${DELIM}*}"

print_status "Login page URL: $LOGIN_PAGE_URL"

# Step 2b: Parse the login form to find action URL and hidden fields
# Dex login form typically has: <form action="..."> and <input type="hidden" name="req" value="AUTH_REQ_ID">
FORM_ACTION=$(echo "$LOGIN_HTML" | grep -oi 'action="[^"]*"' | head -1 | sed "s/action=//i;s/\"//g")

if [ -n "$FORM_ACTION" ]; then
    # Make form action absolute if relative
    if [[ "$FORM_ACTION" == /* ]]; then
        FORM_ACTION="${NETBIRD_URL}${FORM_ACTION}"
    fi
else
    # Fallback: POST to the login page URL itself
    FORM_ACTION="$LOGIN_PAGE_URL"
fi

# Extract hidden 'req' field (Dex auth request ID)
REQ_VALUE=$(echo "$LOGIN_HTML" | grep -oi 'name="req"[^>]*value="[^"]*"' | sed -n 's/.*value="\([^"]*\)".*/\1/p' | head -1)
if [ -z "$REQ_VALUE" ]; then
    # Try reverse attribute order: value before name
    REQ_VALUE=$(echo "$LOGIN_HTML" | grep -oi 'value="[^"]*"[^>]*name="req"' | sed -n 's/.*value="\([^"]*\)".*/\1/p' | head -1)
fi

print_status "Form action: $FORM_ACTION"
print_status "Auth req ID: ${REQ_VALUE:-'(none found)'}"

# Step 2c: Submit login form with credentials and hidden fields
CURL_POST_ARGS=(--data-urlencode "login=${ADMIN_EMAIL}" --data-urlencode "password=${ADMIN_PASSWORD}")
if [ -n "$REQ_VALUE" ]; then
    CURL_POST_ARGS+=(--data-urlencode "req=${REQ_VALUE}")
fi

LOGIN_HEADERS=$(curl -s -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
    -D - -o /dev/null \
    -X POST "$FORM_ACTION" \
    "${CURL_POST_ARGS[@]}" 2>/dev/null || true)

# Get the redirect location from login response
REDIRECT_1=$(echo "$LOGIN_HEADERS" | grep -i "^location:" | head -1 | tr -d '\r\n' | sed 's/^[Ll]ocation: *//')

if [ -z "$REDIRECT_1" ]; then
    print_error "Login form did not return a redirect"
    print_error "Login response headers:"
    echo "$LOGIN_HEADERS" | head -20
    print_error "Login page HTML (first 500 chars):"
    echo "$LOGIN_HTML" | head -c 500
    rm -f "$COOKIE_JAR"
    exit 1
fi

print_status "Login redirect: $REDIRECT_1"

# Make redirect absolute if relative
if [[ "$REDIRECT_1" == /* ]]; then
    REDIRECT_1="${NETBIRD_URL}${REDIRECT_1}"
fi

# Step 2d: Follow redirect chain - Dex approval -> callback with code
# The final redirect goes to localhost:53000 which has no server running
# We need to capture the Location header containing the authorization code
# Use --max-redirs 10 but curl will fail when it tries to connect to localhost:53000
REDIRECT_HEADERS=$(curl -s -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
    -D - -o /dev/null \
    -L --max-redirs 10 "$REDIRECT_1" 2>/dev/null || true)

# Extract the callback URL containing the authorization code
CALLBACK_URL=$(echo "$REDIRECT_HEADERS" | grep -i "^location:" | grep "localhost:53000" | head -1 | tr -d '\r\n' | sed 's/^[Ll]ocation: *//')

if [ -z "$CALLBACK_URL" ]; then
    # Try to get the redirect URL without following it
    CALLBACK_URL=$(curl -s -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
        -o /dev/null -w "%{redirect_url}" \
        "$REDIRECT_1" 2>/dev/null || true)
fi

AUTH_CODE=$(echo "$CALLBACK_URL" | sed -n 's/.*[?&]code=\([^&]*\).*/\1/p')

if [ -z "$AUTH_CODE" ]; then
    print_error "Failed to extract authorization code"
    print_error "Callback URL: $CALLBACK_URL"
    print_error "All redirect headers:"
    echo "$REDIRECT_HEADERS" | grep -i "^location:" || echo "(no location headers)"
    rm -f "$COOKIE_JAR"
    exit 1
fi

print_status "Authorization code obtained"

# Step 2e: Exchange authorization code for JWT tokens
JWT_RESPONSE=$(curl -s -X POST "${NETBIRD_URL}/oauth2/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=authorization_code&code=${AUTH_CODE}&redirect_uri=http://localhost:53000/&client_id=netbird-cli&code_verifier=${CODE_VERIFIER}")

ACCESS_TOKEN=$(echo "$JWT_RESPONSE" | jq -r '.access_token // empty')

rm -f "$COOKIE_JAR"

if [ -z "$ACCESS_TOKEN" ]; then
    print_error "Failed to obtain JWT access token"
    print_error "JWT response: $JWT_RESPONSE"
    exit 1
fi

print_success "JWT access token obtained"

# Step 3: Get user ID if not already known
if [ -z "$USER_ID" ]; then
    print_status "Fetching user ID..."
    USERS_RESPONSE=$(curl -s "${NETBIRD_URL}/api/users" \
        -H "Authorization: Bearer ${ACCESS_TOKEN}")
    USER_ID=$(echo "$USERS_RESPONSE" | jq -r '.[0].id // empty')

    if [ -z "$USER_ID" ]; then
        print_error "Failed to get user ID"
        print_error "Users response: $USERS_RESPONSE"
        exit 1
    fi
fi

print_status "User ID: $USER_ID"

# Step 4: Create a Personal Access Token (PAT)
print_status "Creating Personal Access Token..."
PAT_RESPONSE=$(curl -s -X POST "${NETBIRD_URL}/api/users/${USER_ID}/tokens" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"name":"ci-test","expires_in":365}')

PAT=$(echo "$PAT_RESPONSE" | jq -r '.plain_token // empty')

if [ -z "$PAT" ]; then
    print_error "Failed to create Personal Access Token"
    print_error "PAT response: $PAT_RESPONSE"
    exit 1
fi

print_success "Personal Access Token created successfully"

# Export environment variables
export NETBIRD_API_URL="$NETBIRD_URL"
export NETBIRD_API_TOKEN="$PAT"

# Output for CI (GitHub Actions)
if [ -n "$GITHUB_ENV" ]; then
    echo "NETBIRD_API_URL=${NETBIRD_URL}" >> "$GITHUB_ENV"
    echo "NETBIRD_API_TOKEN=${PAT}" >> "$GITHUB_ENV"
fi

# Output for sourcing in shell
echo ""
echo "# Source these environment variables to run integration tests:"
echo "export NETBIRD_API_URL=${NETBIRD_URL}"
echo "export NETBIRD_API_TOKEN=${PAT}"
