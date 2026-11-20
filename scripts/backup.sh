#!/bin/sh
set -eu

cd "$(dirname "$0")/.."
backup_dir=${1:-backups}
mkdir -p "$backup_dir"
umask 077
backup_file="$backup_dir/bar108-$(date -u +%Y%m%dT%H%M%SZ).dump"
tmp_file="${backup_file}.tmp"
trap 'rm -f "$tmp_file"' EXIT HUP INT TERM

docker compose --env-file .env.production -f compose.production.yml exec -T db \
  sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' > "$tmp_file"
mv "$tmp_file" "$backup_file"
printf '%s\n' "$backup_file"
