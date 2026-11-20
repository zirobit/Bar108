# Operations guide

This guide covers the single-server deployment in [`compose.production.yml`](../compose.production.yml). Read the [English README](../README.md) or [Russian README](../README.ru.md) for setup and API behavior.

## Before first deployment

Choose a host and a domain. Point DNS to the host and allow inbound TCP 80/443. Keep Docker access restricted. Create `.env.production` and the three files under `secrets/` as described in the README. The alert webhook must accept Alertmanager JSON. Verify that backup files can be copied to separate storage and that an external uptime check can reach the public API.

The app applies Goose migrations at startup. There is no automatic rollback of a database migration when an image is rolled back. Review new migration SQL and take a verified backup before releasing it.

## Release

1. Confirm CI is green for the commit to deploy and inspect any new migrations.
2. Run `scripts/backup.sh`, then `scripts/verify-backup.sh PATH_TO_DUMP`. Copy the archive off the host.
3. On the host, update the checkout and run `docker compose --env-file .env.production -f compose.production.yml up -d --build`.
4. Check `docker compose --env-file .env.production -f compose.production.yml ps` and `docker compose --env-file .env.production -f compose.production.yml logs --tail=100 app caddy prometheus alertmanager`.
5. Check `https://YOUR_DOMAIN/ping` from outside the server. Check the local Prometheus targets and alert delivery separately.

If the application fails after a release, inspect logs and the migration applied at startup. Reverting the application image may be safe only if its database schema remains compatible. Restore data only after identifying the exact database and backup; restoration overwrites the selected target's data. Keep the previous archive until the replacement is verified.

## Backup and recovery

`scripts/backup.sh` writes a PostgreSQL custom-format dump under `backups/`. `scripts/verify-backup.sh DUMP` restores into a temporary database and runs a query before removing that database. Schedule both scripts, copy dumps off the host, encrypt them at rest, and define retention. A backup on the same server does not protect against loss of that server.

For an actual recovery, stop application writes, choose a verified archive and a target PostgreSQL instance, then use `pg_restore` according to the target's schema and data state. Practice this in a separate environment before relying on it for production. The verification script intentionally never replaces the production database.

## Health and alerts

`/ping` reports that the HTTP process answers. `/ready` checks PostgreSQL and Redis. Docker uses `/ready` for app health. Prometheus scrapes `/metrics` inside Compose and keeps 15 days of metrics; its UI listens on host loopback port 9090. Caddy blocks public access to `/ready` and `/metrics`. Alertmanager sends availability and repeated HTTP 5xx alerts to the configured webhook. Test delivery with the real receiver before launch. A separate external monitor is needed to detect host or network loss.

## Secrets and access

Keep `.env.production`, `secrets/`, and `backups/` outside Git and restrict filesystem access. Rotate the database password and JWT signing key through an intentional maintenance procedure. Changing the JWT key invalidates existing tokens. Provision the initial admin account through a controlled database operation and audit future role changes. Review the restaurant's data retention and privacy obligations before collecting customer data.
