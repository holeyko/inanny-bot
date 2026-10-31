"""Resolve Compose's env file and adapt URLs; orchestration belongs in Makefile."""
import os
from pathlib import Path
import subprocess
import sys
from urllib.parse import parse_qsl, quote, unquote, urlencode, urlsplit, urlunsplit


class ConfigError(Exception):
    pass


def required(env, name):
    if not env.get(name):
        raise ConfigError(f"Set {name} in ENV_FILE or the shell environment")
    return env[name]


def database(env):
    db = urlsplit(required(env, "DB_URL"))
    if db.scheme not in ("postgres", "postgresql") or not db.hostname or db.fragment:
        raise ConfigError("DB_URL must be a postgres[ql]://user:password@host:port/database URL")
    if not db.username or not db.password or not db.path.strip("/"):
        raise ConfigError("DB_URL requires user, password and database (URL-encode special characters)")
    local = db.hostname.lower() in ("localhost", "127.0.0.1", "::1")
    port = db.port if db.port is not None else 5432
    if not port:
        raise ConfigError("DB_URL port must be between 1 and 65535")
    if local and port != 5432:
        raise ConfigError("Local DB_URL uses the managed Docker database's internal port 5432")
    host = "inanny-postgres" if local else db.hostname
    authority = f"[{host}]" if ":" in host else host
    authority += f":{port}"
    user, password, name = unquote(db.username), unquote(db.password), unquote(db.path[1:])
    if "/" in name:
        raise ConfigError("DB_URL must contain one database name")
    options = parse_qsl(db.query, keep_blank_values=True)
    # JDBC and libpq do not share every option name. Never silently drop one.
    jdbc_names = {"sslmode": "sslmode", "sslrootcert": "sslrootcert", "sslcert": "sslcert",
                  "sslkey": "sslkey", "connect_timeout": "connectTimeout",
                  "application_name": "ApplicationName", "options": "options"}
    if not env.get("MIGRATION_DB_URL"):
        unsupported = [key for key, _ in options if key not in jdbc_names]
        if unsupported:
            raise ConfigError("Unsupported DB_URL option: set an explicit equivalent MIGRATION_DB_URL (JDBC)")
        jdbc_query = urlencode([(jdbc_names[key], value) for key, value in options])
        env["MIGRATION_DB_URL"] = f"jdbc:postgresql://{authority}/{quote(name, safe='')}"
        if jdbc_query:
            env["MIGRATION_DB_URL"] += "?" + jdbc_query
    if not env["MIGRATION_DB_URL"].startswith("jdbc:postgresql://"):
        raise ConfigError("MIGRATION_DB_URL must be a PostgreSQL JDBC URL, without credentials")
    migration = urlsplit(env["MIGRATION_DB_URL"][5:])
    if (migration.username or migration.password or migration.fragment
            or migration.hostname != host or (migration.port if migration.port is not None else 5432) != port
            or unquote(migration.path[1:]) != name):
        raise ConfigError("MIGRATION_DB_URL must target the same resolved host, port and database, without credentials")
    # The existing app constructs postgres://USER:PASSWORD@HOST/NAME. Preserve
    # encoded components and append its URL options to NAME, without Go changes.
    env.update(DB_LOCAL="1" if local else "0", DB_HOST=authority,
               DB_USER=quote(user, safe=""), DB_PASSWORD=quote(password, safe=""),
               DB_NAME=quote(name, safe="") + ("?" + db.query if db.query else ""),
               POSTGRES_USER=user, POSTGRES_PASSWORD=password, POSTGRES_DB=name,
               PGUSER=user, PGPASSWORD=password,
               PGDATABASE=urlunsplit(("postgresql", authority, "/" + quote(name, safe=""), db.query, "")),
               LIQUIBASE_COMMAND_USERNAME=user, LIQUIBASE_COMMAND_PASSWORD=password)


def grafana(env):
    url = urlsplit(required(env, "GRAFANA_URL"))
    if url.scheme not in ("http", "https") or not url.hostname or url.username or url.fragment or url.query:
        raise ConfigError("GRAFANA_URL must be HTTP(S), without credentials, query or fragment")
    local = url.hostname.lower() in ("localhost", "127.0.0.1", "::1")
    port = url.port if url.port is not None else (443 if url.scheme == "https" else 80)
    if not port:
        raise ConfigError("GRAFANA_URL port must be between 1 and 65535")
    env["GRAFANA_LOCAL"] = "1" if local else "0"
    if local:
        required(env, "GRAFANA_ADMIN_PASSWORD")
        env["GRAFANA_PORT"] = str(port)
        if url.scheme != "http" or url.path not in ("", "/"):
            raise ConfigError("Managed local Grafana uses http://localhost:PORT without a subpath")
        env["GRAFANA_HEALTH_URL"] = "http://grafana:3000/api/health"
    else:
        env["GRAFANA_HEALTH_URL"] = env["GRAFANA_URL"].rstrip("/") + "/api/health"


def main():
    env_file, scope, *command = sys.argv[1:]
    if not Path(env_file).exists():
        if env_file != "local/deploy.env":
            raise ConfigError("ENV_FILE does not exist")
        env_file = "/dev/null"
    # Compose handles dotenv quoting/interpolation and shell-over-file precedence.
    # Capture only; never print the resolved environment (it contains secrets).
    resolved = subprocess.run(["docker", "compose", "--env-file", env_file, "-f",
                               "deploy/compose.build.yaml", "config", "--environment"],
                              check=True, capture_output=True, text=True)
    env = dict(os.environ)
    for line in resolved.stdout.splitlines():
        key, sep, value = line.partition("=")
        if not sep or not key.isidentifier():
            raise ConfigError("Deployment env values must be single-line; URL-encode special characters")
        env[key] = value
    env.update(ENV_FILE=env_file, COMPOSE_IGNORE_ORPHANS="true", COMPOSE_REMOVE_ORPHANS="false")
    if scope in ("all", "db"):
        database(env)
    if scope in ("all", "grafana"):
        grafana(env)
    if scope == "all":
        required(env, "TELEGRAM_BOT_TOKEN")
    os.execvpe(command[0], command, env)


if __name__ == "__main__":
    try:
        main()
    except ConfigError as error:
        print(error, file=sys.stderr)
        sys.exit(1)
    except (ValueError, subprocess.CalledProcessError):
        # Parser/subprocess errors may include URLs or credentials.
        print("Invalid deployment configuration or unavailable Compose; see docs/deployment.md", file=sys.stderr)
        sys.exit(1)
