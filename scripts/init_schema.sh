#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

echo "Ensuring database container is running..."
docker compose up -d db

echo "Initializing schema via db container..."
docker compose exec -T db psql -U tex -d tex -v ON_ERROR_STOP=1 \
	< texinroistot-server/internal/db/schema.sql
