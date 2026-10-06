# libreseal — LibreSeal CLI

`libreseal` is the command-line client for [LibreSeal](https://github.com/Dos2Locos/libreseal), a free, self-hosted secrets manager. It authenticates against **your own** LibreSeal server, manages secrets per app and environment, and injects them into processes without writing them to disk.

> LibreSeal CLI is an independent fork of the [Phase CLI](https://github.com/phasehq/cli). It is **not affiliated with or endorsed by Phase / Phi Security Inc.** It has no default cloud host and never downloads installers from third-party infrastructure.

## Install

The CLI is built from source (no pre-built packages are published yet). You need [Go](https://go.dev/dl/) 1.25 or newer.

```sh
git clone https://github.com/Dos2Locos/libreseal-cli.git
cd libreseal-cli
./scripts/install-from-source.sh                 # installs /usr/local/bin/libreseal (sudo if needed)
# or, without sudo:
PREFIX="$HOME/.local" ./scripts/install-from-source.sh
libreseal --version
```

Container image (built locally):

```sh
docker build -t libreseal-cli .
docker run --rm -e LIBRESEAL_HOST -e LIBRESEAL_SERVICE_TOKEN libreseal-cli apps list
```

## Update

`libreseal update` only prints instructions; it never fetches or runs scripts. To update, check out a newer tag and rerun the installer:

```sh
cd libreseal-cli && git fetch --tags
git checkout <tag>
./scripts/install-from-source.sh
```

## Configure

### Interactive users

```sh
libreseal auth                    # prompts for your server URL, then opens the browser (webauth)
libreseal auth --mode token       # paste a personal access token instead
libreseal users whoami
```

### Applications, CI and AI agents (recommended)

Create a **service account** in the LibreSeal UI (Access → Service Accounts) with access only to the apps and environments it needs, generate a token, and provide it through the environment:

```sh
export LIBRESEAL_HOST=https://secrets.example.lan
export LIBRESEAL_SERVICE_TOKEN='pss_service:v2:...'   # never commit or echo this
libreseal apps list
```

| Variable | Purpose |
|----------|---------|
| `LIBRESEAL_HOST` | Server URL (required with a service token; there is no default host) |
| `LIBRESEAL_SERVICE_TOKEN` | Service account (or personal) token for headless use |
| `LIBRESEAL_VERIFY_SSL` | `False` disables TLS verification — only for local test instances with self-signed certificates |
| `LIBRESEAL_OFFLINE` | `1` serves cached data when the server is unreachable |
| `LIBRESEAL_CONFIG_PARENT_DIR_SEARCH_DEPTH` | How many parent directories to search for `.phase.json` |

## Use

```sh
libreseal apps list                                   # apps, IDs and environments
libreseal init --app-id <APP_ID> --env development    # link this directory (.phase.json)

libreseal secrets create DB_PASSWORD --random base64url --length 48 --type sealed
echo "8080" | libreseal secrets create PORT --type config
libreseal secrets list
libreseal secrets get PORT
libreseal secrets update DB_PASSWORD --random hex --length 64
libreseal secrets delete PORT

libreseal secrets import .env --env development
libreseal secrets export --env development --format json

libreseal run 'npm start'                             # inject secrets into a process
libreseal run --env production --tags backend './start.sh'
```

Run `libreseal <command> --help` for all flags.

## AI agents

`libreseal ai enable` (run by a human) installs a skill document for Claude Code, Cursor, Copilot, Codex or OpenCode; `libreseal ai skill` prints it. When the CLI detects an AI agent it blocks `printenv`/`env`/`export`/`set`/`declare`/`compgen` inside `libreseal run` and disables `libreseal shell`. Detection relies on agent-set environment variables and is defence in depth only: give agents a least-privilege service account token and store sensitive values as `sealed`. See also [libreseal-skills](https://github.com/Dos2Locos/libreseal-skills).

## Compatibility with the Phase CLI

| Item | Behaviour |
|------|-----------|
| Command name | `libreseal` (same subcommands and flags as `phase`) |
| Environment variables | `PHASE_HOST`, `PHASE_SERVICE_TOKEN`, `PHASE_VERIFY_SSL`, `PHASE_OFFLINE` and `PHASE_CONFIG_PARENT_DIR_SEARCH_DEPTH` are accepted; `LIBRESEAL_*` takes precedence |
| Project link file | `.phase.json` (unchanged) |
| Local config | `~/.phase/secrets/config.json` and the `phase-cli-user-*` keyring entries are shared with the Phase CLI |
| Default host | **Breaking:** none. `phase` defaulted to Phase Cloud; `libreseal` requires your server URL |
| `update` | **Breaking:** prints instructions instead of running `pkg.phase.dev/install.sh` |
| `dynamic-secrets` | Reports that the feature is unavailable: LibreSeal servers do not implement dynamic secrets |
| Go module path | Still `github.com/phasehq/cli` to keep the diff with upstream small |

The server API and token formats are unchanged. `libreseal` is verified against LibreSeal servers; using it against Phase servers (or `phase` against LibreSeal) is expected to work but is not tested.

## Verified combination

See the [LibreSeal README](https://github.com/Dos2Locos/libreseal#compatibility-and-limitations) for the server / CLI / skills versions verified together.

## Development

```sh
cd src
go vet ./...
go test ./...
go build -o libreseal .
```

## License

GPL-3.0, see [LICENSE](LICENSE). Copyright of the original work belongs to Phase and its contributors; LibreSeal modifications are also released under GPL-3.0.
