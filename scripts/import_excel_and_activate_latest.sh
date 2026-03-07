#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

echo "Running Excel importer in backend container..."
docker compose exec -T backend go run cmd/importer/importer.go

echo "Setting newest version as the only active version..."
docker compose exec -T db psql -U tex -d tex -v ON_ERROR_STOP=1 -c "
UPDATE versions
SET is_active = (
	id = (
		SELECT id
		FROM versions
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	)
);
"

echo "Active version after import:"
docker compose exec -T db psql -U tex -d tex -v ON_ERROR_STOP=1 -c "
SELECT id, created_at, is_active
FROM versions
WHERE is_active = true
ORDER BY created_at DESC, id DESC;
"
