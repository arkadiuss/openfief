#!/usr/bin/env bash
# Tests for scripts/demo/secret.sh (task demo:model-secret and demo:turnstile-secret) against a fake ssh and a fake
# hcloud: no real host, no Hetzner call, no real key (kibi: TEST-tooling-demo-secret). The fake ssh logs its arguments
# and keeps what it reads on standard input, so the tests can see that the value travelled on stdin and nowhere else.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
# A synthetic value that looks like nothing real; every check that must not show it greps for it.
fake_key='sk-or-v1-FAKE-0123456789abcdef-for-tests-only'
fake_turnstile='0xFAKE-turnstile-secret-for-tests-only'
fake_token='fake-hcloud-token-for-tests-only'

mkdir -p "$work/bin" "$work/state"
cat >"$work/bin/hcloud" <<'EOF'
#!/usr/bin/env bash
echo "hcloud $*" >>"$FAKE_LOG"
case "$*" in
  "server ip openfief-demo") echo 192.0.2.10 ;;
  "server ip openfief-gone") echo "hcloud: server not found" >&2; exit 1 ;;
  *) echo "fake hcloud: unexpected call: $*" >&2; exit 2 ;;
esac
EOF
cat >"$work/bin/ssh" <<'EOF'
#!/usr/bin/env bash
printf 'ssh' >>"$FAKE_LOG"
printf ' [%s]' "$@" >>"$FAKE_LOG"
echo >>"$FAKE_LOG"
cat >"$FAKE_STATE/stdin"
if [ -n "${FAKE_SSH_FAIL:-}" ]; then echo "ssh: connect to host 192.0.2.10 port 22: Connection timed out" >&2; exit 255; fi
EOF
chmod +x "$work/bin/hcloud" "$work/bin/ssh"

# run_secret runs secret.sh with the fakes first on PATH, ENV=demo, DEMO_SSH=keys and a token, plus the
# assignments given; its standard input is the file named by STDIN_FILE (empty by default, so no terminal).
run_secret() {
  : >"$work/log"
  rm -f "$work/state/stdin"
  env -i PATH="$work/bin:$PATH" HOME="$work" FAKE_LOG="$work/log" FAKE_STATE="$work/state" \
    ENV=demo DEMO_SSH=keys HCLOUD_TOKEN="$fake_token" "$@" \
    bash "$root/scripts/demo/secret.sh" "$WHICH" <"${STDIN_FILE:-/dev/null}" >"$work/out" 2>&1
}

check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  | /' "$work/out" >&2
    failures=$((failures + 1))
  fi
}

yes_if() {
  if "$@" >/dev/null 2>&1; then echo yes; else echo no; fi
}

# nowhere reports that the value is in none of the run's outputs: the ssh arguments, the hcloud calls (the log), its
# own output, and what it left in the work directory besides the stdin the fake ssh kept.
nowhere() {
  ! grep -rqF -- "$1" "$work/log" "$work/out"
}

# The OpenRouter key, from the environment: created or replaced over SSH, on stdin, in no argument.
WHICH=model
if run_secret OPENROUTER_API_KEY="$fake_key"; then r=yes; else r=no; fi
check "the model key is sent" "$r"
check "the key reaches ssh on standard input, exactly" "$(yes_if test "$(cat "$work/state/stdin")" = "$fake_key")"
check "the key has no trailing newline on stdin" "$(yes_if test "$(wc -c <"$work/state/stdin")" -eq "${#fake_key}")"
check "the key is in no ssh argument and no log" "$(yes_if nowhere "$fake_key")"
check "the key is not printed" "$(yes_if nowhere "$fake_key")"
check "ssh goes to root at the machine's address" "$(yes_if grep -qF '[root@192.0.2.10]' "$work/log")"
check "ssh does not ask for a password" "$(yes_if grep -qF '[BatchMode=yes]' "$work/log")"
want="[k3s kubectl -n openfief create secret generic openrouter--api-key --from-file=api-key=/dev/stdin --dry-run=client -o yaml | k3s kubectl apply -f -]"
check "the remote command creates or replaces the Secret openrouter--api-key" "$(yes_if grep -qF -- "$want" "$work/log")"
check "the machine is found by its environment's label" "$(yes_if grep -qxF 'hcloud server ip openfief-demo' "$work/log")"
check "the result names the Secret and no value" "$(yes_if grep -q 'Secret openrouter--api-key is created or replaced' "$work/out")"

# A pipe, for a script that has no terminal: the first line is the key.
printf '%s\n' "$fake_key" >"$work/stdin.txt"
if STDIN_FILE="$work/stdin.txt" run_secret; then r=yes; else r=no; fi
check "a key piped in is sent" "$r"
check "the piped key reaches ssh without its newline" "$(yes_if test "$(cat "$work/state/stdin")" = "$fake_key")"
check "the piped key is printed nowhere" "$(yes_if nowhere "$fake_key")"

# Rotation is a second run: the same command, the new value.
if run_secret OPENROUTER_API_KEY="${fake_key}-rotated"; then r=yes; else r=no; fi
check "running it again sends the new value" "$(yes_if test "$(cat "$work/state/stdin")" = "${fake_key}-rotated")"
check "the rotation is the same remote command" "$(yes_if grep -qF -- "$want" "$work/log")"

# The Turnstile secret key: another Secret and file, the same way.
WHICH=turnstile
if run_secret TURNSTILE_SECRET_KEY="$fake_turnstile"; then r=yes; else r=no; fi
check "the Turnstile secret key is sent" "$r"
check "it travels on standard input" "$(yes_if test "$(cat "$work/state/stdin")" = "$fake_turnstile")"
want="[k3s kubectl -n openfief create secret generic ui--turnstile-secret --from-file=secret-key=/dev/stdin --dry-run=client -o yaml | k3s kubectl apply -f -]"
check "the remote command names ui--turnstile-secret" "$(yes_if grep -qF -- "$want" "$work/log")"
check "the Turnstile key is printed nowhere" "$(yes_if nowhere "$fake_turnstile")"
check "the model's variable is not what the Turnstile command reads" \
  "$(if run_secret OPENROUTER_API_KEY="$fake_key"; then echo no; else echo yes; fi)"

# The Grafana Cloud credentials: one Secret, three keys, in namespace observability, sent as an env file on stdin.
WHICH=grafana
fake_endpoint='https://otlp-gateway-prod-eu-west-0.grafana.net/otlp'
fake_instance='1234567'
fake_gtoken='glc_FAKE0123456789abcdef-for-tests-only='
if run_secret GRAFANA_OTLP_ENDPOINT="$fake_endpoint" GRAFANA_INSTANCE_ID="$fake_instance" GRAFANA_OTLP_TOKEN="$fake_gtoken"; then r=yes; else r=no; fi
check "the Grafana credentials are sent" "$r"
printf 'endpoint=%s\ninstance-id=%s\ntoken=%s\n' "$fake_endpoint" "$fake_instance" "$fake_gtoken" >"$work/want-env"
check "they reach ssh on standard input as an env file, exactly" "$(yes_if cmp -s "$work/want-env" "$work/state/stdin")"
for v in "$fake_gtoken" "$fake_endpoint" "$fake_instance"; do
  check "a Grafana value is in no ssh argument, log or output" "$(yes_if nowhere "$v")"
done
want="[k3s kubectl -n observability create secret generic grafana-otlp --from-env-file=/dev/stdin --dry-run=client -o yaml | k3s kubectl apply -f - && { k3s kubectl -n observability rollout restart daemonset/otel-collector || true; }]"
check "the remote command creates or replaces grafana-otlp in observability and restarts the collector" "$(yes_if grep -qF -- "$want" "$work/log")"
check "the result names the Secret and no value" "$(yes_if grep -q 'Secret grafana-otlp is created or replaced in the namespace observability' "$work/out")"
if run_secret GRAFANA_OTLP_ENDPOINT="$fake_endpoint/" GRAFANA_INSTANCE_ID="$fake_instance" GRAFANA_OTLP_TOKEN="$fake_gtoken"; then r=yes; else r=no; fi
check "a trailing slash on the endpoint is dropped" "$(yes_if cmp -s "$work/want-env" "$work/state/stdin")"
# A pipe: the three values, one per line, endpoint first.
printf '%s\n%s\n%s\n' "$fake_endpoint" "$fake_instance" "$fake_gtoken" >"$work/stdin.txt"
if STDIN_FILE="$work/stdin.txt" run_secret; then r=yes; else r=no; fi
check "three values piped in are sent" "$r"
check "the piped values reach ssh as the same env file" "$(yes_if cmp -s "$work/want-env" "$work/state/stdin")"
# What the endpoint, the instance ID and the token must look like; each refusal sends nothing and prints no value.
refuse_grafana() {
  local name="$1"
  shift
  if run_secret "$@"; then r=no; else r=yes; fi
  check "$name is refused" "$r"
  check "$name sends nothing" "$(yes_if test ! -e "$work/state/stdin")"
  check "$name prints no value" "$(yes_if nowhere "$fake_gtoken")"
}
refuse_grafana "an http endpoint" GRAFANA_OTLP_ENDPOINT="http://otlp-gateway-prod-eu-west-0.grafana.net/otlp" GRAFANA_INSTANCE_ID="$fake_instance" GRAFANA_OTLP_TOKEN="$fake_gtoken"
refuse_grafana "an endpoint on another host" GRAFANA_OTLP_ENDPOINT="https://otlp.example.com/otlp" GRAFANA_INSTANCE_ID="$fake_instance" GRAFANA_OTLP_TOKEN="$fake_gtoken"
refuse_grafana "an endpoint that only starts like Grafana's" GRAFANA_OTLP_ENDPOINT="https://otlp.grafana.net.example.com/otlp" GRAFANA_INSTANCE_ID="$fake_instance" GRAFANA_OTLP_TOKEN="$fake_gtoken"
refuse_grafana "an endpoint with credentials in it" GRAFANA_OTLP_ENDPOINT="https://user:pw@otlp-gateway-prod-eu-west-0.grafana.net/otlp" GRAFANA_INSTANCE_ID="$fake_instance" GRAFANA_OTLP_TOKEN="$fake_gtoken"
refuse_grafana "an instance ID that is no number" GRAFANA_OTLP_ENDPOINT="$fake_endpoint" GRAFANA_INSTANCE_ID="abc" GRAFANA_OTLP_TOKEN="$fake_gtoken"
refuse_grafana "a token with white space" GRAFANA_OTLP_ENDPOINT="$fake_endpoint" GRAFANA_INSTANCE_ID="$fake_instance" GRAFANA_OTLP_TOKEN=$'two words'
refuse_grafana "a token that is too short" GRAFANA_OTLP_ENDPOINT="$fake_endpoint" GRAFANA_INSTANCE_ID="$fake_instance" GRAFANA_OTLP_TOKEN="abc"
refuse_grafana "no token" GRAFANA_OTLP_ENDPOINT="$fake_endpoint" GRAFANA_INSTANCE_ID="$fake_instance"
check "no token is refused naming the variable" "$(yes_if grep -q 'GRAFANA_OTLP_TOKEN' "$work/out")"
refuse_grafana "no SSH" DEMO_SSH=off GRAFANA_OTLP_ENDPOINT="$fake_endpoint" GRAFANA_INSTANCE_ID="$fake_instance" GRAFANA_OTLP_TOKEN="$fake_gtoken"
check "the refusal for no SSH says what to run" "$(yes_if grep -q 'DEMO_SSH=keys' "$work/out")"
if run_secret GRAFANA_OTLP_ENDPOINT="$fake_endpoint" GRAFANA_INSTANCE_ID="$fake_instance" GRAFANA_OTLP_TOKEN="$fake_gtoken" FAKE_SSH_FAIL=1; then r=no; else r=yes; fi
check "a failing ssh fails the Grafana task" "$r"
check "and prints no value and does not claim success" "$(yes_if bash -c "! grep -q 'created or replaced' '$work/out' && ! grep -qF -- '$fake_gtoken' '$work/out'")"

# Refusals send nothing and call nothing.
WHICH=model
if run_secret OPENROUTER_API_KEY="$fake_key" DEMO_SSH=off; then r=no; else r=yes; fi
check "without DEMO_SSH=keys it refuses" "$r"
check "the refusal says what to run" "$(yes_if grep -q 'DEMO_SSH=keys' "$work/out")"
check "the refusal makes no ssh call" "$(yes_if test ! -e "$work/state/stdin")"
if run_secret OPENROUTER_API_KEY="$fake_key" DEMO_SSH=; then r=no; else r=yes; fi
check "an empty DEMO_SSH refuses" "$r"
if run_secret; then r=no; else r=yes; fi
check "no key and no terminal refuses" "$r"
check "the refusal for no key names the variable" "$(yes_if grep -q 'OPENROUTER_API_KEY' "$work/out")"
check "no key makes no ssh call" "$(yes_if test ! -e "$work/state/stdin")"
if run_secret OPENROUTER_API_KEY=$'two words'; then r=no; else r=yes; fi
check "a value with white space is refused" "$r"
check "it makes no ssh call and prints no value" "$(yes_if bash -c "test ! -e '$work/state/stdin'")"
if run_secret OPENROUTER_API_KEY="$fake_key" HCLOUD_TOKEN=; then r=no; else r=yes; fi
check "without a Hetzner token it refuses" "$r"
check "that refusal names HCLOUD_TOKEN and prints no value" "$(yes_if bash -c "grep -q HCLOUD_TOKEN '$work/out' && ! grep -qF -- '$fake_key' '$work/out'")"
if run_secret OPENROUTER_API_KEY="$fake_key" ENV=Demo; then r=no; else r=yes; fi
check "a malformed ENV is refused" "$r"
if run_secret OPENROUTER_API_KEY="$fake_key" ENV=gone; then r=no; else r=yes; fi
check "a machine that does not exist is refused" "$r"
check "that refusal sends nothing" "$(yes_if test ! -e "$work/state/stdin")"
WHICH=other
if run_secret OPENROUTER_API_KEY="$fake_key"; then r=no; else r=yes; fi
check "an unknown secret is refused" "$r"
WHICH=model

# A failing ssh fails the command, and still prints no value.
if run_secret OPENROUTER_API_KEY="$fake_key" FAKE_SSH_FAIL=1; then r=no; else r=yes; fi
check "a failing ssh fails the task" "$r"
check "a failing ssh does not claim success" "$(yes_if bash -c "! grep -q 'created or replaced' '$work/out'")"
check "a failing ssh prints no value" "$(yes_if nowhere "$fake_key")"

if [ "$failures" -ne 0 ]; then
  echo "demo-secret: $failures of $cases checks failed" >&2
  exit 1
fi
echo "demo-secret: $cases checks passed"
