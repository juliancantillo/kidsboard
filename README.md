# Kidsboard

RPG-style household activity tracker for kids. Parents log chores, school work, prayers, meals — kids see their character sheet: per-category XP bars, unlocked achievements with humorously-named titles, a spendable points balance, and a shop of rewards. Single Go binary, embedded SQLite, server-rendered Tailwind UI with a pixel-art aesthetic.

![kidsboard](assets/screenshot.png)

## Quick start — install on a Kubernetes cluster

The simplest path: pull the packaged Helm chart from the GitHub Release and install. Works on any Kubernetes 1.22+ cluster.

```bash
helm install kidsboard \
  https://github.com/juliancantillo/kidsboard/releases/download/v0.1.0/kidsboard-0.1.0.tgz \
  --namespace kidsboard --create-namespace
```

Then port-forward and open it:

```bash
kubectl -n kidsboard port-forward svc/kidsboard 8080:80
open http://localhost:8080
```

First time only — seed the curated categories, activity types, and achievements:

```bash
kubectl -n kidsboard exec -it kidsboard-0 -- /usr/local/bin/kidsboard seed
```

(Or set `app.seedOnDeploy: true` in the chart to run it on every rollout.)

Parents reach `/admin` to create kids and log activities. Kids land on `/` and pick their character.

> **GHCR visibility** — the very first time you cut a release, the image lands at `ghcr.io/<you>/kidsboard` as a **private** package. Until you flip it to public at `https://github.com/users/<you>/packages/container/kidsboard/settings`, the cluster won't be able to pull it without an `imagePullSecret`. After flipping, the chart works out of the box.

## Local cluster recipes

Pick whichever you already have running.

### kind

```bash
# Spin up a cluster (one-time)
kind create cluster --name kidsboard

# Install
helm install kidsboard \
  https://github.com/juliancantillo/kidsboard/releases/download/v0.1.0/kidsboard-0.1.0.tgz \
  --namespace kidsboard --create-namespace

# Reach it
kubectl -n kidsboard port-forward svc/kidsboard 8080:80
```

### k3d

```bash
k3d cluster create kidsboard --port "8080:80@loadbalancer"

helm install kidsboard \
  https://github.com/juliancantillo/kidsboard/releases/download/v0.1.0/kidsboard-0.1.0.tgz \
  --namespace kidsboard --create-namespace \
  --set ingress.enabled=true \
  --set 'ingress.hosts[0].host=localhost' \
  --set 'ingress.hosts[0].paths[0].path=/' \
  --set 'ingress.hosts[0].paths[0].pathType=Prefix'

open http://localhost:8080
```

### Docker Desktop Kubernetes

Enable Kubernetes in Docker Desktop → Settings → Kubernetes, then:

```bash
kubectl config use-context docker-desktop

helm install kidsboard \
  https://github.com/juliancantillo/kidsboard/releases/download/v0.1.0/kidsboard-0.1.0.tgz \
  --namespace kidsboard --create-namespace

kubectl -n kidsboard port-forward svc/kidsboard 8080:80
```

### Raspberry Pi 4/5 with k3s

The container image is multi-arch (`linux/amd64` + `linux/arm64`); the Pi pulls the arm64 variant automatically.

```bash
# On the Pi — install k3s if not already there
curl -sfL https://get.k3s.io | sh -

# From your laptop (with KUBECONFIG pointed at the Pi)
helm install kidsboard \
  https://github.com/juliancantillo/kidsboard/releases/download/v0.1.0/kidsboard-0.1.0.tgz \
  --namespace kidsboard --create-namespace \
  --set persistence.storageClass=local-path  # k3s default
```

Expose on the home network (the Pi's IP, no port-forward needed):

```bash
helm upgrade kidsboard \
  https://github.com/juliancantillo/kidsboard/releases/download/v0.1.0/kidsboard-0.1.0.tgz \
  --namespace kidsboard \
  --set service.type=LoadBalancer
```

## Configuration

Every flag is also a `KIDSBOARD_*` environment variable. In the Helm chart, override via `--set app.<key>=<value>`.

| Helm value | Env var | Default | Purpose |
|---|---|---|---|
| `app.addr` | `KIDSBOARD_ADDR` | `:8080` | HTTP listen address |
| `app.dbPath` | `KIDSBOARD_DB` | `/data/kidsboard.db` | SQLite file path |
| `app.shutdownTimeout` | `KIDSBOARD_SHUTDOWN_TIMEOUT` | `30s` | Grace period to drain in-flight requests |
| `app.readTimeout` | `KIDSBOARD_READ_TIMEOUT` | `30s` | HTTP read timeout |
| `app.idleTimeout` | `KIDSBOARD_IDLE_TIMEOUT` | `60s` | Keep-alive idle timeout |
| `app.logLevel` | `KIDSBOARD_LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `app.logFormat` | `KIDSBOARD_LOG_FORMAT` | `json` | `json` (prod) or `text` (dev) |
| `app.seedOnDeploy` | — | `false` | Run `kidsboard seed` as a second init container |
| `persistence.size` | — | `1Gi` | PVC size for SQLite |
| `persistence.storageClass` | — | (default) | Storage class name |
| `image.tag` | — | `.Chart.AppVersion` | Override to pin a specific version |
| `ingress.enabled` | — | `false` | Expose via Ingress |

## Surfaces

- `GET /` — kid selector. Pick a character to land on a profile.
- `GET /kids/{id}` — character sheet: per-category levels with progress bars, earned achievements with titles, next-3 closest achievements, points balance, recent activity.
- `GET /admin` — parent panel: create kids, log activities. **No link from `/`** — parents bookmark the URL.
- `GET /admin/achievements` — achievement CRUD (multi-rule combinator forms).
- `GET /healthz` and `GET /readyz` — liveness / readiness probes.

## Upgrading

The PVC survives `helm upgrade`. Schema migrations run automatically in the `migrate` init container on every rollout (idempotent — already-applied migrations are no-ops).

```bash
helm upgrade kidsboard \
  https://github.com/juliancantillo/kidsboard/releases/download/v0.1.1/kidsboard-0.1.1.tgz \
  --namespace kidsboard
```

## Local development

```bash
make check          # gofmt + go test ./... + go build ./...
make seed           # populate a local kidsboard.db with curated content
make run            # go run . serve --db kidsboard.db --addr :8080
make docker         # build for your host arch
make docker-pi      # cross-build linux/arm64 image, loaded into local Docker
make helm-template  # render the chart to stdout
```

Run with env-var-driven config (12-factor):

```bash
KIDSBOARD_DB=/tmp/kidsboard.db \
KIDSBOARD_ADDR=:8888 \
KIDSBOARD_LOG_FORMAT=text \
KIDSBOARD_LOG_LEVEL=debug \
  go run . serve
```

## Repository layout

```
.
├── cmd/                  # cobra entry points (serve, migrate, seed, healthcheck, version)
├── internal/
│   ├── domain/           # pure Go structs — no tags, no behavior
│   ├── storage/          # SQLite open + WAL + goose migrations + sqlc DBTX
│   │   ├── migrations/   # goose .sql files (schema source of truth)
│   │   └── sqldb/        # sqlc-generated (private)
│   ├── repository/       # interface + sqlite impl, maps sqlc rows → domain
│   ├── service/          # services (Kid, Category, Activity, Balance, Achievement, Profile…)
│   ├── http/             # router + controllers
│   ├── view/             # html/template renderer + Tailwind UI + embedded static assets
│   ├── applog/           # log/slog setup
│   └── seed/             # curated Spanish seed data (categories, types, achievements)
├── queries/              # sqlc .sql input
├── deploy/helm/kidsboard # Helm chart
├── .github/workflows/    # CI + Release pipelines
├── Dockerfile            # multi-arch (amd64 + arm64) build
└── Makefile              # check, seed, run, docker, helm-* targets
```

## Architecture highlights

- **Single binary, embedded everything** — SQLite, migrations, html templates, pixel-art assets all embedded into the Go binary via `embed.FS`. No assets to ship alongside.
- **Achievement engine** — multi-rule composite achievements with `ALL`/`ANY` combinator. Metrics: `count`, `xp`, `points` (category-scoped or global), `level`. Fixed-point iteration handles cascading unlocks where one achievement's bonus points satisfy another's threshold.
- **Append-only event tables** — `activities`, `kid_achievements`, `redemptions`, `point_adjustments` are append-only with `voided_at` flags. Soft-deletable config tables (`archived_at`). No DELETE in the app code.
- **Single-writer StatefulSet** — SQLite is single-writer, so the chart hard-locks `replicas: 1`. The PVC survives upgrades.

## Acknowledgments

Pixel art assets from [Craftpix](https://craftpix.net/) — fairy avatar icons, RPG loot icons, mine-location icons, and nature backgrounds. License per their [free license terms](https://craftpix.net/file-licenses/).

## License

MIT. See `LICENSE`.
