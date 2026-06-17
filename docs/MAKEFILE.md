# Makefile Reference

The [Makefile](../Makefile) is the main entry point for local development, Docker orchestration, database setup, and running services. Run `make help` to print all targets.

## Compose profiles

| Profile | Services | Use case |
|---------|----------|----------|
| `infra` | `db`, `migrate` | Database only |
| `full` | `db`, `migrate`, `control`, `market`, `gateway` | Complete stack |

Variables used internally:

- `COMPOSE` = `docker compose --profile full`
- `COMPOSE_INFRA` = `docker compose --profile infra`

---

## Setup targets

| Target | Description |
|--------|-------------|
| `make setup` | Alias for `setup-local` |
| `make setup-local` | Copies `.env.local.example` → `.env` (if missing), then runs `db-init` against local/Homebrew MySQL |
| `make setup-docker` | Copies `.env.docker.example` → `.env`, checks/installs Docker Desktop |
| `make env-local` | Only creates `.env` from local template |
| `make env-docker` | Overwrites `.env` from Docker template |
| `make install-docker` | Installs Docker Desktop via Homebrew (macOS) if `docker` is missing |

---

## Database targets

| Target | Description |
|--------|-------------|
| `make db-init` | Creates DB if needed, applies pending migrations via `scripts/init-db.sh local` |
| `make db-init-docker` | Runs `migrate up` inside Docker (DB container must be running) |
| `make db-reset` | **Destructive.** Stops stack, deletes volumes, starts DB, reapplies all migrations |
| `make migrate-up` | Apply pending migrations via Docker (`migrate up`) |
| `make migrate-down` | Roll back the last migration via Docker |
| `make migrate-status` | Show goose migration status |
| `make migrate-version` | Print current goose version number |

See [MIGRATIONS.md](./MIGRATIONS.md) for how migrations work.

**Note:** `migrate-*` targets run inside the `migrate` container so they work when `.env` has `DB_HOST=db`. Ensure MySQL is running first (`make up-db` or `make up-full`).

---

## Docker stack targets

| Target | Description |
|--------|-------------|
| `make up-db` | Start MySQL + run migrations (`infra` profile) |
| `make up-full` | Start DB, migrate, then all app services (`full` profile). **Does not rebuild images.** |
| `make up-full-build` | **Rebuild all images**, then `up-full`. Use after code changes. |
| `make up` | Alias for `up-full` |
| `make down` | Stop all containers. Data persists in `spice_ledger_mysql_data` volume. |
| `make down-volumes` | Stop containers **and delete** the MySQL volume |
| `make build` | Build all Docker images without starting |
| `make logs` | Tail logs for full-profile services |
| `make ps` | List all compose containers |

### Rebuild rule of thumb

| Situation | Command |
|-----------|---------|
| First time / fresh clone | `make setup-docker && make up-full-build` |
| Code changed (Go, Dockerfiles) | `make up-full-build` or `docker compose --profile full build <service>` |
| Only `.env` or migrations changed | `make up-full` or `make migrate-up` |
| GraphQL/REST behavior not updating | Rebuild gateway — `make up-full` alone reuses old images |

---

## Local service targets

Run these in **separate terminals** (or use VS Code **Full Stack (local)**). Requires `.env` with `DB_HOST=localhost` and migrations applied.

| Target | Service | Default port |
|--------|---------|--------------|
| `make run-control` | Control gRPC | 50051 |
| `make run-market` | Market gRPC | 50052 |
| `make run-gateway` | Unified HTTP (REST + GraphQL) | 8080 |

Startup order: `control` → `market` → `gateway`.

---

## Quality targets

| Target | Description |
|--------|-------------|
| `make test` | `go test ./...` |
| `make lint` | `go build ./...` (compile check) |
| `make clean` | Clear Go cache and remove local binary artifacts |

---

## Common workflows

### Local dev (Homebrew MySQL)

```bash
brew services start mysql
make setup
make run-control    # terminal 1
make run-market     # terminal 2
make run-gateway    # terminal 3
```

### Full Docker stack

```bash
make setup-docker
open -a Docker      # wait until Docker is running
make up-full-build
```

### Docker DB + native Go services

```bash
make up-db
# Set .env: DB_HOST=127.0.0.1, DB_HOST_PORT=3306 (if 3306 is taken by Homebrew MySQL)
make run-control
# ... other run-* targets
```

### After pulling new migrations

```bash
make migrate-up     # Docker DB
# or
make db-init        # local DB
```
