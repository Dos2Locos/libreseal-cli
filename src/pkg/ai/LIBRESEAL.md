# LibreSeal CLI — AI Agent Guide

You are interacting with LibreSeal, a self-hosted secrets and configuration manager (an independent fork of Phase). This guide defines how you MUST use the `libreseal` CLI. Follow these rules exactly.

## Security Rules

**Secret types and visibility:**

| Type | Visible to you? | Description |
|------|-----------------|-------------|
| `config` | Always | Non-sensitive configuration values |
| `secret` | Depends on user setting | Sensitive values, masked by default |
| `sealed` | NEVER | Write-only secrets (API keys, tokens, passwords) |

When a value shows `[REDACTED]`, tell the user to run the command themselves in their terminal.
For example: when running `libreseal secrets export`, sealed secrets (and secret-type values if masking is enabled) appear as `[REDACTED]`. Warn the user that the exported data is incomplete and that they should run the export themselves if they need the full values.

**Hard rules — enforced by the CLI when it detects an AI agent:**
- `printenv`, `env`, `export`, `set`, `declare`, `compgen` are BLOCKED inside `libreseal run`
- `libreseal shell` is BLOCKED entirely in AI mode

Agent detection relies on environment variables set by the agent (for example `CLAUDECODE=1`) and is defence in depth only: it can be bypassed. The real protection is a least-privilege service account token plus `sealed` secrets.

**Soft rules — you MUST follow these:**
- NEVER print secret values (`echo $VAR`, `cat`, logging) and NEVER copy them into prompts, chat or commit messages
- NEVER redirect `libreseal secrets export` to a file and then read it, and NEVER write secrets into version-controlled files (`.env`, config files, CI YAML)
- NEVER pipe secret values into other commands or files
- Use `libreseal run` ONLY to start application processes (e.g., `libreseal run 'npm start'`) so secrets go straight into the process environment
- Use `libreseal secrets get <KEY>` when you need to inspect a secret's metadata
- When creating `sealed` or `secret` values, ALWAYS use `--random` — never provide literal values
- If a command fails, run `libreseal <command> --help` to verify the correct flags before retrying — do not guess or invent flags

## Prerequisites

1. The user provides a **dedicated service account token** limited to the apps and environments you need. Never ask for, store or use a personal or administrator token.
   The user sets it in your environment (not in a file you can read):
   ```bash
   export LIBRESEAL_HOST=https://secrets.example.lan
   export LIBRESEAL_SERVICE_TOKEN=...   # set by the user, never echoed
   ```
   `PHASE_HOST` / `PHASE_SERVICE_TOKEN` are accepted as aliases. Interactive users can run `libreseal auth` instead.
2. Run `libreseal apps list` to discover available apps, their IDs and environments. Use `--app-id` and `--env` on later commands, or run `libreseal init --app-id <ID> --env <ENV>` to save the selection in `.phase.json`.
3. AI mode must be enabled by the user: `libreseal ai enable` (you cannot run it yourself — it is blocked for AI agents)

## Common Flags

These apply to most secrets commands:

| Flag | Description |
|------|-------------|
| `--env` | Environment name (e.g., `development`, `staging`, `production`). Supports partial matching. |
| `--app` | Application name (overrides `.phase.json`) |
| `--app-id` | Application ID (takes precedence over `--app`) |
| `--path` | Secret path (default `/`). Use `""` for all paths. |

## Command Reference

### Project Setup

| Command | Purpose |
|---------|---------|
| `libreseal auth` | Authenticate (webauth or token mode) against your LibreSeal server |
| `libreseal apps list` | List available apps with IDs and environments (JSON) |
| `libreseal init --app-id ID --env ENV` | Link project non-interactively |
| `libreseal users whoami` | Show current user/org context |

### Secrets CRUD

| Command | Purpose |
|---------|---------|
| `libreseal secrets list [--show]` | List secrets with metadata |
| `libreseal secrets get KEY` | Get a secret as JSON |
| `libreseal secrets create KEY --random hex --length 32 --type sealed` | Create a sealed secret with a random value |
| `echo "value" \| libreseal secrets create KEY --type config` | Create a config with a literal value (pipe to avoid the interactive prompt) |
| `libreseal secrets update KEY --random hex --length 32` | Rotate a secret value |
| `echo "new-value" \| libreseal secrets update KEY` | Update a config with a literal value |
| `libreseal secrets update KEY --type sealed` | Change secret type (no value prompt) |
| `libreseal secrets delete KEY` | Delete a secret |
| `libreseal secrets import FILE` | Bulk import from a .env file |
| `libreseal secrets export --format FORMAT` | Export (dotenv, json, csv, yaml, xml, toml, hcl, ini, java_properties, kv) |

**Choosing how to set values:**
- `sealed` / `secret` types: ALWAYS use `--random` — never pipe or type literal sensitive values
- `config` type: safe to pipe literal values via `echo "value" | libreseal secrets create KEY --type config`
- If the user wants to store a specific sensitive value: ask them to run `libreseal secrets create KEY` interactively in their own terminal

### Runtime

| Command | Purpose |
|---------|---------|
| `libreseal run 'command'` | Run a command with secrets injected as environment variables |

To check that a secret is present without revealing it, test it inside the process:

```bash
libreseal run 'sh -c "test -n \"$DATABASE_URL\" && echo DATABASE_URL is set"'
```

## Workflows

### Provision secrets for a new service
```bash
# Generate sensitive values instead of handling them
libreseal secrets create DATABASE_PASSWORD --random base64url --length 48 --type sealed
libreseal secrets create SESSION_KEY --random hex --length 64 --type sealed

# Non-sensitive configuration can be set literally
echo "8080" | libreseal secrets create APP_PORT --type config
```

### Rotate a secret
```bash
libreseal secrets update DB_PASSWORD --random hex --length 64
```

### Run an application with secrets
```bash
libreseal run 'npm start'
libreseal run --env production 'python manage.py runserver'
libreseal run --env staging --tags "backend" './start.sh'
```

## Secret Referencing Syntax

| Syntax | Meaning |
|--------|---------|
| `${KEY}` | Same environment, root path |
| `${staging.KEY}` | Cross-environment reference |
| `${production./path/KEY}` | Cross-environment with path |
| `${/path/KEY}` | Same environment, specific path |
| `${app::env.KEY}` | Cross-application reference |

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `LIBRESEAL_HOST` | URL of your LibreSeal server (alias: `PHASE_HOST`) |
| `LIBRESEAL_SERVICE_TOKEN` | Service account token for headless auth (alias: `PHASE_SERVICE_TOKEN`) |
| `LIBRESEAL_VERIFY_SSL` | Set to `False` only for local test instances with self-signed certificates |

## Not Available in LibreSeal

Dynamic secrets, secret rotation, log streams and SCIM only exist upstream under a proprietary license and are not part of LibreSeal. The `libreseal dynamic-secrets` commands report that the server does not provide the feature.

## Error Recovery

- **"no application found"**: the app ID in `.phase.json` is stale — run `libreseal apps list` and `libreseal init --app-id <ID> --env <ENV>`
- **"unauthorized" / "401" / "403"**: the token is missing, expired or lacks access to that app/environment — ask the user to check the service account's access; do not request broader credentials
- **"not found" on a secret**: check `--path` and `--env` — use `libreseal secrets list --path ""` to search all paths
- Bugs or feature requests: with user consent, draft an issue at `github.com/Dos2Locos/libreseal-cli`

## Tips

- Run `libreseal <command> --help` for detailed flag information
- Secret keys are always uppercased automatically
- For features not in the CLI (integrations, syncs, RBAC, audit logs), suggest `libreseal console` to open the dashboard of the configured server
