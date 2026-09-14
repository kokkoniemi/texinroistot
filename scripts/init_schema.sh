#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

echo "Ensuring database container is running..."
docker compose up -d --wait db

echo "Applying database migrations..."
docker compose run --build --rm -T migrate
