#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTAINER_ENGINE="${CONTAINER_ENGINE:-docker}"
CHECK_DIR="$(mktemp -d)"
CHECK_ID="$(basename "${CHECK_DIR}" | tr '[:upper:]' '[:lower:]')"
IMAGE="texinroistot-migration-check:${CHECK_ID}"
CONTAINER=""

cleanup() {
	if [[ -n "${CONTAINER}" ]]; then
		"${CONTAINER_ENGINE}" rm -fv "${CONTAINER}" >/dev/null
	fi
	"${CONTAINER_ENGINE}" image rm "${IMAGE}" >/dev/null 2>&1 || true
	rm -rf "${CHECK_DIR}"
}
trap cleanup EXIT

tar -C "${ROOT_DIR}/texinroistot-server" -cf - Dockerfile.migrator internal/db/migrations \
	| "${CONTAINER_ENGINE}" build -f Dockerfile.migrator -t "${IMAGE}" -
CONTAINER="$("${CONTAINER_ENGINE}" run -d --tmpfs /var/lib/postgresql/data \
	-e POSTGRES_USER=migration_check -e POSTGRES_PASSWORD=migration_check \
	-e POSTGRES_DB=migration_check docker.io/library/postgres:17-alpine)"

ready=false
for attempt in {1..60}; do
	if "${CONTAINER_ENGINE}" exec "${CONTAINER}" pg_isready -h 127.0.0.1 -U migration_check -d migration_check >/dev/null 2>&1; then
		ready=true
		break
	fi
	sleep 1
done
if [[ "${ready}" != true ]]; then
	echo "Disposable PostgreSQL did not become ready" >&2
	exit 1
fi

sql() {
	"${CONTAINER_ENGINE}" exec -i "${CONTAINER}" psql -X -U migration_check -d migration_check -v ON_ERROR_STOP=1 "$@"
}

migrate() {
	"${CONTAINER_ENGINE}" run --rm --network "container:${CONTAINER}" "${IMAGE}" \
		-database "postgres://migration_check:migration_check@127.0.0.1:5432/migration_check?sslmode=disable" "$@"
}

schema() {
	"${CONTAINER_ENGINE}" exec "${CONTAINER}" pg_dump -U migration_check -d "$1" \
		--schema-only --no-owner --no-privileges --exclude-table=public.schema_migrations \
		| sed '/^\\restrict /d; /^\\unrestrict /d'
}

assert_version() {
	local expected_file expected_version state
	local migration_files=("${ROOT_DIR}/texinroistot-server/internal/db/migrations/"*.up.sql)
	expected_file="${migration_files[-1]##*/}"
	expected_version="${expected_file%%_*}"
	state="$(sql -Atc "SELECT version || ':' || dirty FROM public.schema_migrations")"
	if [[ "${state}" != "$((10#${expected_version})):false" ]]; then
		echo "Unexpected migration state: ${state}" >&2
		exit 1
	fi
	migrate version
}

echo "Applying migrations and checking schema objects..."
migrate up
assert_version
sql < "${ROOT_DIR}/texinroistot-server/internal/db/check_migrations.sql"
schema migration_check > "${CHECK_DIR}/first.sql"

echo "Verifying the reviewed schema snapshot..."
"${CONTAINER_ENGINE}" exec "${CONTAINER}" createdb -U migration_check snapshot_check
"${CONTAINER_ENGINE}" exec -i "${CONTAINER}" psql -X -U migration_check -d snapshot_check -v ON_ERROR_STOP=1 \
	< "${ROOT_DIR}/texinroistot-server/internal/db/schema.sql"
schema snapshot_check > "${CHECK_DIR}/snapshot.sql"
diff -u "${CHECK_DIR}/snapshot.sql" "${CHECK_DIR}/first.sql"

echo "Verifying repeated up is a no-op..."
migrate up
assert_version
schema migration_check > "${CHECK_DIR}/repeated.sql"
diff -u "${CHECK_DIR}/first.sql" "${CHECK_DIR}/repeated.sql"

echo "Reversing migrations on the disposable database..."
migrate down -all
remaining="$(sql -Atc "SELECT
	(SELECT count(*) FROM pg_class WHERE relnamespace = 'public'::regnamespace
	 AND relkind IN ('r', 'p', 'S', 'v', 'm', 'f') AND relname <> 'schema_migrations') +
	(SELECT count(*) FROM pg_type WHERE typnamespace = 'public'::regnamespace AND typtype = 'e') +
	(SELECT count(*) FROM public.schema_migrations)")"
if [[ "${remaining}" != 0 ]]; then
	echo "Down migrations left application objects or migration state behind" >&2
	exit 1
fi

echo "Applying migrations again..."
migrate up
assert_version
sql < "${ROOT_DIR}/texinroistot-server/internal/db/check_migrations.sql"
schema migration_check > "${CHECK_DIR}/roundtrip.sql"
diff -u "${CHECK_DIR}/first.sql" "${CHECK_DIR}/roundtrip.sql"

echo "Checking rejection of a legacy database with existing tables..."
if "${CONTAINER_ENGINE}" run --rm --network "container:${CONTAINER}" "${IMAGE}" \
	-database "postgres://migration_check:migration_check@127.0.0.1:5432/snapshot_check?sslmode=disable" up \
	> "${CHECK_DIR}/legacy-error.txt" 2>&1; then
	echo "Migrations unexpectedly accepted an unversioned populated schema" >&2
	exit 1
fi
legacy_dirty="$("${CONTAINER_ENGINE}" exec "${CONTAINER}" psql -X -U migration_check -d snapshot_check -Atc \
	'SELECT dirty FROM public.schema_migrations')"
if [[ "${legacy_dirty}" != t ]]; then
	echo "Failed legacy migration did not record dirty state" >&2
	exit 1
fi
schema snapshot_check > "${CHECK_DIR}/legacy.sql"
diff -u "${CHECK_DIR}/snapshot.sql" "${CHECK_DIR}/legacy.sql"

echo "PostgreSQL 17 migration up/down/up, schema parity, and legacy rejection passed."
