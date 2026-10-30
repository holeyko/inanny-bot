# Deployment

Pushes to `main` and manual workflow runs use `.github/workflows/deploy.yaml`.
The single deployment job checks out once, builds the commit-tagged image,
ensures PostgreSQL is ready, applies migrations, and replaces the bot. The
self-hosted runner must be the intended production Docker host.

The production concurrency group prevents overlapping deployments. Active runs
are not cancelled, so migrations and container replacement can finish safely.
The old bot remains running until the build and migrations succeed.

## Build cache

The Dockerfile requires BuildKit and keeps Go modules and compiled packages in
cache mounts. The workflow installs Buildx with `docker/setup-buildx-action`,
using a named `docker-container` builder with `keep-state: true` so caches survive
job cleanup. `make docker-build` explicitly uses `docker buildx build --load`:
it cannot fall back to the legacy builder, and the resulting image is loaded into
the local Docker daemon for `make docker-run`. Local builds also require the
Docker Buildx plugin (`docker buildx version` should succeed).

Reuse the same builder between deployments. `.dockerignore` keeps
local configuration, Git metadata, and build artifacts out of the build context;
only module files, `cmd/`, and `internal/` are copied into the builder.

## Startup verification

`make docker-run` calls `make docker-ready`. It waits up to
`BOT_STARTUP_TIMEOUT` seconds (default: 60), failing if the container exits,
restarts, or never emits the successful startup-check marker.

The application emits that marker only after checking its own database
connection and successfully making a Telegram polling request. These probes
share a 10-second deadline. The probe uses offset zero and does not acknowledge
pending updates; normal polling still handles them. Failed probes do not disable
the application's existing degraded-mode operation, but deployment is reported
as failed. This is a startup check, not continuous health monitoring.

There is no automatic rollback after replacement failure. Inspect the failure
and redeploy a known-good commit if necessary, accounting for applied migrations.

## Failure diagnostics

The workflow calls `make docker-diagnostics` only after failure. Local readiness
failures also invoke it. It prints selected state, mounts, and networks for the
two application containers, plus at most 200 recent log lines per container
(`--since 10m`). It never dumps container environment variables.

Docker log rotation is separate: limiting diagnostic output does not limit log
storage on the host. Configure retention on the production Docker daemon as
needed.
