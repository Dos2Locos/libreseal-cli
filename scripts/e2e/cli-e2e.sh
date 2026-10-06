#!/bin/bash
# LibreSeal CLI end-to-end test against a running LibreSeal server.
#
# Intended to run in a throwaway Linux container (no system keyring) that
# shares the network namespace of the LibreSeal nginx container, e.g.:
#   docker run --rm --network container:libreseal-nginx \
#     -v "$PWD/libreseal:/opt/libreseal:ro" -v "$PWD/scripts/e2e/cli-e2e.sh:/t/run.sh:ro" \
#     -v "$RW_TOKEN_FILE:/run/tok-rw:ro" -v "$RO_TOKEN_FILE:/run/tok-ro:ro" -e APP=<app-id> \
#     alpine:3.21 sh -c 'apk add -q bash expect ca-certificates && bash /t/run.sh'
# /run/tok-rw: service account token with read/write on APP/development only.
# /run/tok-ro: service account token with read-only access to APP/development.
# Only synthetic secrets are created; values are never printed.
set -u
export HOME=/tmp/home; mkdir -p "$HOME" /work; cd /work || exit 1
export LIBRESEAL_VERIFY_SSL=False
L=/opt/libreseal
APP="${APP:?set APP to the application ID}"
pass=0; fail=0
ok(){ echo "PASS: $1"; pass=$((pass+1)); }
ko(){ echo "FAIL: $1"; fail=$((fail+1)); }
EXPECTED="synthetic-$(head -c 12 /dev/urandom | od -An -tx1 | tr -d ' \n')"
FIXVAL="synthetic-$(head -c 8 /dev/urandom | od -An -tx1 | tr -d ' \n')"

$L --version >/dev/null && $L --help | grep -q LibreSeal && ok "libreseal --help shows LibreSeal" || ko "help"

# 1. auth --mode token (interactive prompts driven by expect)
cat > /tmp/auth.exp <<'EXP'
set timeout 60
set tok [string trim [read [open "/run/tok-rw"]]]
spawn /opt/libreseal auth --mode token
expect "server URL" { send "https://localhost\r" }
expect "Token" { send "$tok\r" }
expect eof
catch wait result
exit [lindex $result 3]
EXP
expect /tmp/auth.exp > /tmp/auth.log 2>&1 && ok "auth --mode token against https://localhost" || { ko "auth"; tail -5 /tmp/auth.log; }
$L users whoami > /tmp/whoami.log 2>&1 && ok "users whoami" || { ko whoami; cat /tmp/whoami.log; }
$L apps list 2>&1 | grep -q '"demo"\|demo' && ok "apps list shows demo" || ko "apps list"

# 2. init + CRUD
$L init --app-id $APP --env development >/dev/null 2>&1 && test -f .phase.json && ok "init writes .phase.json" || ko init
$L secrets create LS_CLI_SEALED --random hex --length 32 --type sealed >/dev/null 2>&1 && ok "create sealed random" || ko "create sealed"
echo "$EXPECTED" | $L secrets create LS_RUN_KEY --type secret >/dev/null 2>&1 && ok "create secret from stdin" || ko "create secret"
echo "info" | $L secrets create LS_CLI_CONFIG --type config >/dev/null 2>&1 && ok "create config" || ko "create config"
$L secrets list > /tmp/list.log 2>&1; grep -q LS_CLI_SEALED /tmp/list.log && grep -q LS_RUN_KEY /tmp/list.log && ok "secrets list" || ko "list"
$L secrets get LS_CLI_CONFIG 2>&1 | grep -q '"info"' && ok "get config value" || ko "get config"
$L ai enable --mask --path /tmp/skill/libreseal-cli/SKILL.md >/dev/null 2>&1 && grep -q "libreseal-cli-skill-version" /tmp/skill/libreseal-cli/SKILL.md && ok "ai enable installs LibreSeal skill" || ko "ai enable"
CLAUDECODE=1 $L secrets get LS_CLI_SEALED 2>&1 | grep -q REDACTED && ok "sealed value redacted for agents" || ko "sealed redaction"
CLAUDECODE=1 $L ai enable --mask >/dev/null 2>&1 && ko "agent could run ai enable" || ok "ai enable refused for agents"
echo "debug" | $L secrets update LS_CLI_CONFIG >/dev/null 2>&1 && $L secrets get LS_CLI_CONFIG 2>&1 | grep -q '"debug"' && ok "update" || ko "update"
$L secrets delete LS_CLI_CONFIG >/dev/null 2>&1 && ! $L secrets list 2>&1 | grep -q LS_CLI_CONFIG && ok "delete" || ko "delete"

# 3. import / export
printf 'LS_IMPORT_A=%s\nLS_IMPORT_B=plain\n' "$FIXVAL" > fixture.env
$L secrets import fixture.env --env development >/dev/null 2>&1 && ok "import .env" || ko "import"
$L secrets export --env development --format json > /tmp/export.json 2>/dev/null
grep -q LS_IMPORT_A /tmp/export.json && grep -q LS_IMPORT_B /tmp/export.json && ok "export json contains imported keys" || ko "export"
rm -f fixture.env /tmp/export.json

# 4. run injection without exposing value
OUT=$($L run --env development "sh -c 'test \"\$LS_RUN_KEY\" = \"$EXPECTED\" && echo MATCH'" 2>&1)
echo "$OUT" | grep -q MATCH && ok "run injects secret (MATCH)" || ko "run inject"
echo "$OUT" | grep -q "$EXPECTED" && ko "run output leaked value" || ok "run output does not contain the value"
$L run 'true' >/dev/null 2>&1 && ok "run 'true' exits 0 (no lease errors)" || ko "run true"

# 5. dynamic secrets unavailable
DS=$($L dynamic-secrets list --env development 2>&1); echo "$DS" | grep -q "not available on this server" && ok "dynamic-secrets reports unavailable" || { ko "dynamic msg"; echo "$DS"; }

# 6. update: no Phase infrastructure
$L update | grep -q "Dos2Locos/libreseal-cli" && ! $L update | grep -q phase.dev && ok "update prints LibreSeal instructions" || ko update

# 7. env-var auth, aliases, least privilege
mv $HOME/.phase $HOME/.phase.bak
RO_TOKEN="$(cat /run/tok-ro)"; RW_TOKEN="$(cat /run/tok-rw)"
export LIBRESEAL_HOST=https://localhost LIBRESEAL_SERVICE_TOKEN="$RO_TOKEN"
$L secrets list --app-id $APP --env development >/dev/null 2>&1 && ok "read-only token can list via LIBRESEAL_* env" || ko "ro list"
if echo x | $L secrets create LS_DENIED --app-id $APP --env development > /tmp/deny.log 2>&1; then ko "read-only create should fail"; else ok "read-only create denied ($(grep -oiE 'permission|403|forbidden' /tmp/deny.log | head -1))"; fi
$L secrets list --app-id $APP --env production >/dev/null 2>&1 && ko "production should be denied" || ok "production denied for scoped token"
unset LIBRESEAL_HOST LIBRESEAL_SERVICE_TOKEN
export PHASE_HOST=https://localhost PHASE_SERVICE_TOKEN="$RO_TOKEN"
$L secrets list --app-id $APP --env development >/dev/null 2>&1 && ok "PHASE_* aliases work" || ko "aliases"
unset PHASE_HOST
NOHOST=$($L secrets list --app-id $APP --env development 2>&1); echo "$NOHOST" | grep -q LIBRESEAL_HOST && ok "no default host: error asks for LIBRESEAL_HOST" || { ko "no-host"; echo "$NOHOST"; }
export PHASE_HOST=https://localhost

# 8. agent mode guard
CLAUDECODE=1 $L run --app-id $APP --env development 'printenv' > /tmp/agent.log 2>&1 && ko "printenv not blocked" || { grep -q "blocked" /tmp/agent.log && ok "printenv blocked in agent mode" || ko "agent block msg"; }
grep -q "$EXPECTED" /tmp/agent.log && ko "agent log leaked" || true
CLAUDECODE=1 $L shell >/dev/null 2>&1 && ko "shell not blocked" || ok "shell blocked in agent mode"

# cleanup test secrets with rw token
export PHASE_SERVICE_TOKEN="$RW_TOKEN"
$L secrets delete LS_CLI_SEALED LS_RUN_KEY LS_IMPORT_A LS_IMPORT_B --app-id $APP --env development >/dev/null 2>&1
echo "TOTAL: pass=$pass fail=$fail"
[ $fail -eq 0 ]
