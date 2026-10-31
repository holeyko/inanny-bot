# Deployment

All deployment orchestration is visible in the root `Makefile`; there are no
deployment shell scripts. The three independently operable runtime stacks are:

| Compose file | Services | Preserved project identity |
| --- | --- | --- |
| `deploy/bot/docker-compose.yaml` | bot | `inanny`, service `bot` |
| `deploy/postgresql/docker-compose.yaml` | PostgreSQL, one-shot Liquibase and DB probe (`tools` profile) | `inanny`, service `postgres` |
| `deploy/grafana/docker-compose.yaml` | Grafana, Loki, Alloy **together** | `inanny-observability` |

`deploy/compose.build.yaml` is only a credential-free image build fragment, not
a fourth runtime stack. `deploy/Dockerfile` builds `cmd/inanny/main.go` and runs
`/app/main`. `deploy/config.py` is a small Python standard-library URL adapter:
it asks Compose to resolve the env file, validates URLs, translates their
components into the existing application's DB environment and JDBC settings,
then execs Make. It contains no deployment stages or service orchestration.

## Configuration and placement

Requires Docker Engine, Compose v2, Buildx and Python 3 on the intended bot Docker
host. Execute from this checkout **on that host**, directly or through SSH.
Bind mounts refer to that host's checkout. Do not point a laptop checkout at a
remote Docker context: configuration, SQL and certificate mounts would be wrong.

```sh
mkdir -p local
cp deploy/.env.example local/deploy.env
chmod 600 local/deploy.env
# Edit DB_URL, GRAFANA_URL, Telegram token and local Grafana admin password.
make compose-check
make deploy
```

`ENV_FILE=/absolute/path/to/file` selects another file; absent default
`local/deploy.env` means shell-only settings. Explicitly missing files fail.
Exported shell values override file values, using Compose's own dotenv parsing.
Use single-line values; percent-encode special characters in URL components.
Never print resolved Compose config or `config --environment`: both may contain
secrets. `make compose-check` uses `config --quiet` instead.

The canonical settings are:

* `DB_URL=postgresql://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=...`
* `GRAFANA_URL=http[s]://HOST:PORT[/optional-remote-subpath]`
* `TELEGRAM_BOT_TOKEN`, plus any optional AI settings in `deploy/.env.example`.

Placement uses the **parsed exact hostname**, not a string prefix. `localhost`,
`127.0.0.1` and `::1` mean the local Docker host. A hostname such as
`localhost.example.com` is remote. URLs cannot contain fragments; Grafana URLs
cannot contain credentials or queries. There are no fake runtime credentials.

### Local dependencies

Local `DB_URL` selects the managed `inanny-postgres` container, on internal port
5432; another localhost DB port is rejected rather than silently targeting the
wrong database. The DB is not published on the host. Localhost in this URL is a
**placement selector**, not loopback inside the bot. It becomes
`inanny-postgres:5432` for bot, authenticated probe and Liquibase.
An unrelated host PostgreSQL instance is not automatically adopted.

Local `GRAFANA_URL` must be `http://localhost:PORT` (or another exact loopback
host), without a subpath. Its port controls the loopback-only published Grafana
port; container probes use `http://grafana:3000`. This starts **all three** of
Grafana, Loki and Alloy on the bot host. `GRAFANA_ADMIN_PASSWORD` is required;
`GRAFANA_ADMIN_USER` defaults to `admin`. Existing Grafana passwords are not
reset by changing init environment variables.

`up --no-recreate` leaves existing database and logging containers intact,
including when an image or config has changed. It starts stopped managed
containers and creates absent ones. Upgrades, changed ports/configurations and
credential rotation require explicit maintenance, not an implicit deployment.
The existing external `inanny-postgres-data` volume, `inanny-network`, container
names and Compose service labels are preserved. Named Grafana/Loki/Alloy volumes
retain their historical `inanny-observability_*` identities after the file move.
Never change PostgreSQL major versions against an existing data volume without
a planned upgrade. Set `POSTGRES_IMAGE` to the current database major version.

### Remote dependencies

Non-loopback `DB_URL` connects directly over `inanny-network` to the external
database, preserving its port, encoded credentials, database name and libpq URL
options. No local PostgreSQL is started or modified. Allow network access from
the bot Docker host, provision the database/user in advance and use suitable TLS.

The existing Go app constructs a URL from `DB_HOST`, `DB_USER`, `DB_PASSWORD`,
`DB_NAME`. The adapter keeps the user/password/database URL-encoded, includes
the port in `DB_HOST`, and appends the unchanged query to `DB_NAME`; it does not
alter the application's Go DB code. PostgreSQL initialization and Liquibase use
the decoded credentials instead. Do not pass the derived app variables as raw
PostgreSQL initialization settings.

Automatic JDBC translation supports `sslmode`, `sslrootcert`, `sslcert`,
`sslkey`, `connect_timeout` -> `connectTimeout`, `application_name` ->
`ApplicationName`, and `options`. Any other option requires an explicit
`MIGRATION_DB_URL=jdbc:postgresql://HOST:PORT/DATABASE?...` with equivalent JDBC
options; nothing is silently discarded. The override must target the same DB
and security policy; credentials are passed separately. Client TLS formats can
differ between libpq/pgx and JDBC (especially private keys): provision compatible
files and reference the appropriate ones in each URL. Set `DB_TLS_DIR` to an
absolute host directory; it is mounted read-only at `/run/db-tls` for bot, DB
probe and migration. URL certificate paths must be container paths, not host
paths. With no certificates, the default mount is ignored local scratch storage.

Non-loopback `GRAFANA_URL` means the **entire Grafana/Loki/Alloy stack is remote**.
No local logging component is deployed, including Alloy. The remote stack must
already be provisioned and reachable from the bot network. Deployment probes
`GRAFANA_URL/api/health`; provide optional `GRAFANA_HEALTH_TOKEN` for a bearer-auth
reverse proxy. It does not use or require the remote Grafana admin password.
The URL must expose Grafana's health API, not an HTML login/redirect endpoint.
Remote Loki/Alloy health, storage and ingestion are the remote operator's
responsibility: a healthy Grafana API alone does **not** prove log ingestion.

#### Remote log collection requires an explicit secure source

Alloy on another host cannot discover the bot through its own Docker socket.
Merely setting `GRAFANA_URL` does **not** deliver bot logs remotely. No `LOKI_URL`
is needed by the bot, no local collector is installed and no unauthenticated
Docker TCP port is exposed.

One supported source configuration is an operator-managed SSH **Unix-socket**
tunnel on the logging host, forwarding to the bot host's Docker socket:

```sh
# Run on the logging host, under a restricted operational account.
# Provision /run/inanny with restrictive ownership first. No public TCP listener.
ssh -N -o ExitOnForwardFailure=yes -o StreamLocalBindUnlink=yes \
  -L /run/inanny/bot-docker.sock:/var/run/docker.sock bot-logs@bot-host
```

Provision SSH authentication, strict host-key verification, socket ownership
and supervised tunnel restart separately. Access to Docker's API is effectively
root-equivalent, even through a read-only socket bind: trust Alloy and restrict
the account, key and socket. Mount the forwarded socket into **remote** Alloy
by setting `ALLOY_DOCKER_SOCKET=/run/inanny/bot-docker.sock` on the logging host
before creating that stack. Both `discovery.docker` and `loki.source.docker` in
the provided Alloy configuration use the mounted socket. Recreate the Alloy
container after a tunnel socket inode is replaced; a bind mount can retain the
old inode. This is an explicit operator-managed transport, not an automatically
provisioned or health-verified feature of `make deploy`.

To provision the whole remote stack using this checkout, run on the logging
host with `GRAFANA_URL=http://localhost:3000`, its admin credentials and forwarded
socket configuration: `make logging-up`. On the bot host use its **remote**
Grafana URL. The optional `local/logs` file source refers to the logging host's
files, not files on the bot host. A single forwarded socket collects one Docker
host; collection from multiple Docker hosts requires explicit additional remote
Alloy sources/configuration. Until the source is configured and tested, remote
bot ingestion is **not operational**.

## Ordered deployment

`make deploy` (aliases `docker-deploy`, `docker-bar`, `docker-run`) performs:

0. Validate selected runtime config, reject unmanaged DB/bot containers and
   ensure `inanny-network` exists. Remote dependencies are not adopted.
1. **In parallel**, build and load the bot image, ensure local PostgreSQL if
   selected, and ensure the entire local Grafana stack if selected.
2. **In parallel**, inspect Docker network liveness, run an authenticated
   `SELECT 1` from a one-shot PostgreSQL client on the bot/migration network,
   and check Grafana HTTP health. A local stack also checks Loki `/ready` and
   Alloy `/-/ready`. DB and HTTP probes have bounded retries/timeouts.
3. Run one-shot Liquibase migration on that network, against the selected DB.
4. Force-replace **only** the bot using the loaded image, then verify startup.

Recursive `make -j3` invocations provide actual parallel siblings even with
plain `make deploy`. Separate invocations are stage barriers. A failed stage
stops the pipeline (already running siblings finish); the old bot is untouched
until the build, probes and migration all succeed. Network creation tolerates
parallel initial creators. The startup check waits `BOT_STARTUP_TIMEOUT` seconds
(default 60), requiring `running 0` and the app's successful DB/Telegram startup
marker. It is not continuous health monitoring. There is no automatic rollback
after replacement, and applied schema migrations may constrain manual rollback.

Useful independent commands:

* `make docker-build`: credential-free image build; no runtime services.
* `make db-bootstrap`: ensure selected DB/network, then authenticated probe.
* `make db-migrate`: bootstrap/probe then migrate; no bot/logging changes.
* `make logging-up`: ensure selected whole stack; remote selection is a no-op.
* `make logging-status`, `make docker-ready`, `make docker-diagnostics`.
* `make docker-clean`: explicitly remove only the owned bot.

The bot and PostgreSQL stacks intentionally retain the same historical project
identity. Orphan removal is forcibly disabled. Never run project-wide `down`,
`down -v` or `--remove-orphans`: the other independent stack is not an orphan to
delete. Bare Compose commands bypass URL adaptation and ordered deployment;
use the Make entrypoints.

## Legacy containers: explicit one-time maintenance

Compose cannot adopt containers by adding labels. Legacy/wrong-project
`inanny-bot` or local `inanny-postgres` causes a refusal, with no automatic
removal or data migration. Plan an operator-controlled outage:

1. Inspect the existing DB image/version, credentials and data mounts privately.
   Confirm `inanny-postgres-data` is mounted at `/var/lib/postgresql/data`; if
   not, stop and create a separate data-migration plan. Take and verify a backup.
   Keep the current PostgreSQL major and credentials in deployment config.
2. Prepare image/config first (`make docker-build`, `make compose-check`).
3. Explicitly stop and rename the legacy containers (skip absent containers):

   ```sh
   docker stop inanny-bot inanny-postgres
   docker rename inanny-bot inanny-bot-legacy
   docker rename inanny-postgres inanny-postgres-legacy
   docker update --restart=no inanny-bot-legacy inanny-postgres-legacy
   ```

4. Run `make deploy`, verify the DB/application/logs and retain backups and old
   containers until verified. Never run both databases against the same volume.

Rollback is manual: stop/remove only new bot/DB containers, restore old names
and restart policies, then start the old database before the bot. Schema changes
may require restoring a backup; renaming containers is not a database rollback.
Never delete the database volume as part of this procedure.

## Logging, CI and cache

All published logging ports are loopback-only. Access local Grafana with
`ssh -N -L 3000:127.0.0.1:3000 user@docker-host`, then browse
`http://127.0.0.1:3000`. Grafana provisions Loki and the bot dashboard. Explore
queries include `{container="inanny-bot"}` and `{container="inanny-postgres"}`.
Alloy persists offsets, Loki stores seven days on its volume, and Grafana
persists settings. Docker JSON logs rotate at 10 MB x 3 files; a collector outage
can lose logs after rotation. This storage is not an off-host backup.

`.github/workflows/deploy.yaml` calls the same `make deploy` on the intended
self-hosted bot Docker host. Required secrets: `DB_URL`, `TELEGRAM_BOT_TOKEN`,
and `GRAFANA_ADMIN_PASSWORD` only for local Grafana. Configure `GRAFANA_URL` as a
repository variable; optional TLS/JDBC, health bearer and AI settings appear in
the workflow/example. CI uses `ENV_FILE=/dev/null`, never a leftover host env
file. Its concurrency group serializes deployments and does not cancel running
migrations. Operators must also avoid overlapping manual runs.

BuildKit caches Go modules/build output. The workflow uses a named Buildx
`docker-container` builder with `keep-state: true`. Make resolves only the
credential-free build fragment and passes it to `docker buildx bake -f - --load
bot`, explicitly loading the image into the deployment daemon. No credentials
or runtime services are needed: `make docker-build ENV_FILE=/dev/null` works.
The runtime bot references the same `inanny-bot:${IMAGE_VERSION:-latest}` tag.
`.dockerignore` excludes local config, Git metadata and build artifacts.

Failure diagnostics print selected state/mount/network fields (never Config.Env)
and at most 200 recent log lines per application container from the last ten
minutes. Logs can contain sensitive user data; restrict access. No runtime
health/remote ingestion is guaranteed by config validation or an image build.
