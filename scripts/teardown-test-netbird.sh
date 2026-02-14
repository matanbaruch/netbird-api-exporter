#!/bin/bash
set -e

BLUE='\033[0;34m'
GREEN='\033[0;32m'
NC='\033[0m'

print_status() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

COMPOSE_FILE="tests/docker-compose.netbird.yml"

print_status "Stopping self-hosted NetBird server..."
docker compose -f "$COMPOSE_FILE" down -v

print_success "NetBird test server stopped and volumes removed"
