#!/bin/sh
set -eu

if [ "$#" -ne 1 ] || [ ! -f "$1" ]; then
  echo "usage: scripts/verify-backup.sh BACKUP_FILE" >&2
  exit 2
fi

backup_file=$(cd "$(dirname "$1")" && pwd)/$(basename "$1")
cd "$(dirname "$0")/.."
scratch_db="bar108_restore_$$"
compose() {
  docker compose --env-file .env.production -f compose.production.yml "$@"
}
cleanup() {
  compose exec -T db sh -c 'dropdb --if-exists -U "$POSTGRES_USER" "$1"' sh "$scratch_db" >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM

compose exec -T db sh -c 'createdb -U "$POSTGRES_USER" "$1"' sh "$scratch_db"
compose exec -T db sh -c 'pg_restore --exit-on-error -U "$POSTGRES_USER" -d "$1"' sh "$scratch_db" < "$backup_file"
compose exec -T db sh -c 'psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -Atc "SELECT count(*) FROM users"' sh "$scratch_db" >/dev/null
echo "Backup restored successfully to a temporary database."
