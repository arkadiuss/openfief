#!/usr/bin/env bash
# Tests for scripts/demo/demo.sh against a fake hcloud with synthetic prices and
# inventories; no Hetzner API call (kibi: TEST-tooling-demo). Needs gojq on PATH (task test adds .bin).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
token='fake-token-for-tests-only'

# The fake hcloud logs each call (one line of arguments) and answers from the fixture files named in FAKE_*.
mkdir -p "$work/bin"
cat >"$work/bin/hcloud" <<'EOF'
#!/usr/bin/env bash
echo "$*" >>"$FAKE_LOG"
case "$*" in
  "all list --selector app=openfief --output json") cat "$FAKE_INVENTORY" ;;
  "all list --selector app=openfief,env="*) cat "$FAKE_ENV_INVENTORY" ;;
  "api /pricing") cat "$FAKE_PRICING" ;;
  "ssh-key list --selector openfief-ssh=true --output json") cat "$FAKE_KEYS" ;;
  "firewall describe "*) echo '{"id": 7}' ;;
  "server ip "*) echo 192.0.2.10 ;;
  "firewall create "* | "firewall replace-rules "*)
    while [ $# -gt 0 ]; do
      if [ "$1" = --rules-file ]; then cp "$2" "$FAKE_STATE/rules.json"; fi
      shift
    done ;;
  "server create "*)
    while [ $# -gt 0 ]; do
      if [ "$1" = --user-data-from-file ]; then cp "$2" "$FAKE_STATE/user-data.yaml"; fi
      shift
    done ;;
  *" delete "* | "firewall apply-to-resource "*) ;;
  *) echo "fake hcloud: unexpected call: $*" >&2; exit 2 ;;
esac
EOF
chmod +x "$work/bin/hcloud"

# Synthetic prices: gross = net x 1.19.
cat >"$work/pricing.json" <<'EOF'
{"pricing": {"currency": "EUR", "vat_rate": "19.00",
  "image": {"price_per_gb_month": {"net": "0.0110000000", "gross": "0.0130900000"}},
  "volume": {"price_per_gb_month": {"net": "0.0440000000", "gross": "0.0523600000"}},
  "server_backup": {"percentage": "20.00"},
  "floating_ips": [{"type": "ipv4", "prices": [{"location": "nbg1", "price_monthly": {"net": "3.0000000000", "gross": "3.5700000000"}}]}],
  "primary_ips": [{"type": "ipv4", "prices": [{"location": "nbg1", "price_hourly": {"net": "0.0008", "gross": "0.0010"},
    "price_monthly": {"net": "0.5000000000", "gross": "0.5950000000"}}]}],
  "server_types": [
    {"name": "cax11", "prices": [{"location": "nbg1", "price_monthly": {"net": "6.4900000000", "gross": "7.7231000000"}}]},
    {"name": "cax21", "prices": [{"location": "nbg1", "price_monthly": {"net": "10.9900000000", "gross": "13.0781000000"}}]},
    {"name": "cx23", "prices": [{"location": "nbg1", "price_monthly": {"net": "5.9900000000", "gross": "7.1281000000"}}]}],
  "load_balancer_types": []}}
EOF
echo '{"servers": [], "firewalls": [], "primary_ips": [], "volumes": []}' >"$work/empty.json"
cat >"$work/existing.json" <<'EOF'
{"servers": [{"id": 41, "name": "openfief-demo", "status": "running", "labels": {"app": "openfief", "env": "demo"},
  "server_type": {"name": "cax11"}, "location": {"name": "nbg1"}, "backup_window": null,
  "public_net": {"ipv4": {"ip": "192.0.2.10"}, "firewalls": [{"id": 7, "status": "applied"}]}}],
 "firewalls": [{"id": 7, "name": "openfief-demo", "labels": {"app": "openfief", "env": "demo"}, "rules": []}],
 "primary_ips": [{"id": 5, "name": "auto-ip", "type": "ipv4", "assignee_id": 41, "location": {"name": "nbg1"}}]}
EOF
cat >"$work/other-env.json" <<'EOF'
{"servers": [{"id": 42, "name": "openfief-other", "status": "running", "labels": {"app": "openfief", "env": "other"},
  "server_type": {"name": "cax11"}, "location": {"name": "nbg1"}, "backup_window": null,
  "public_net": {"ipv4": {"ip": "192.0.2.11"}, "firewalls": []}}],
 "firewalls": [],
 "primary_ips": [{"id": 6, "name": "spare", "type": "ipv4", "assignee_id": null, "location": {"name": "nbg1"}}]}
EOF
echo '[{"id": 1, "name": "alice", "labels": {"openfief-ssh": "true"}}]' >"$work/keys.json"
echo '[]' >"$work/no-keys.json"

# run_demo runs demo.sh with the fake hcloud and fresh state; output lands in $work/out, calls in $work/log.
run_demo() {
  rm -rf "$work/state" "$work/log"
  mkdir -p "$work/state"
  : >"$work/log"
  env PATH="$work/bin:$PATH" HCLOUD_TOKEN="$token" FAKE_LOG="$work/log" FAKE_STATE="$work/state" \
    FAKE_PRICING="$work/pricing.json" FAKE_INVENTORY="${INVENTORY:-$work/empty.json}" \
    FAKE_ENV_INVENTORY="${ENV_INVENTORY:-$work/existing.json}" FAKE_KEYS="${KEYS:-$work/keys.json}" \
    K3S_VERSION=v1.36.5+k3s1 K3S_SHA256_ARM64="$(printf 'a%.0s' {1..64})" K3S_SHA256_AMD64="$(printf 'b%.0s' {1..64})" \
    K3S_INSTALL_SHA256="$(printf 'c%.0s' {1..64})" FLUX_VERSION=v2.9.6 FLUX_INSTALL_SHA256="$(printf 'd%.0s' {1..64})" \
    "$@" bash "$root/scripts/demo/demo.sh" "${ACTION:-plan}" >"$work/out" 2>&1
}

# check records a failed expectation with the case name and the run's output.
check() {
  local name="$1" ok="$2"
  cases=$((cases + 1))
  if [ "$ok" != yes ]; then
    echo "FAIL: $name" >&2
    sed 's/^/  | /' "$work/out" >&2
    failures=$((failures + 1))
  fi
}

# yes_if prints yes when the command succeeds, so expectations read as one line each.
yes_if() {
  if "$@" >/dev/null 2>&1; then echo yes; else echo no; fi
}

# The token is required, and the error names it without printing a value.
if ACTION=plan run_demo ENV=demo HCLOUD_TOKEN=; then r=no; else r=yes; fi
check "missing token is refused" "$r"
check "the refusal names HCLOUD_TOKEN" "$(yes_if grep -q 'HCLOUD_TOKEN is not set' "$work/out")"
check "missing token makes no API call" "$(yes_if test ! -s "$work/log")"

# plan with nothing in the project: one new cax11 and its IPv4, within budget, and nothing created.
if ACTION=plan run_demo ENV=demo; then r=yes; else r=no; fi
check "plan within budget succeeds" "$r"
check "plan shows the server price" "$(yes_if grep -Eq 'new server openfief-demo \(cax11, nbg1\) +7\.7231 EUR' "$work/out")"
check "plan shows the IPv4 price" "$(yes_if grep -Eq 'new IPv4 of openfief-demo +0\.5950 EUR' "$work/out")"
check "plan shows the total" "$(yes_if grep -Eq 'total +8\.3181 EUR per month, gross' "$work/out")"
check "plan creates nothing" "$(yes_if bash -c "! grep -Eq 'create|delete|replace-rules' '$work/log'")"

# The budget refuses a bigger server, and counts what other demos already cost.
if ACTION=up run_demo ENV=demo DEMO_SERVER_TYPE=cax21; then r=no; else r=yes; fi
check "up over budget is refused" "$r"
check "refusal shows the total and budget" "$(yes_if grep -q 'refusing: 13.6731 EUR per month would exceed the budget of 10.0000 EUR' "$work/out")"
check "refused up creates nothing" "$(yes_if bash -c "! grep -q create '$work/log'")"
if ACTION=up INVENTORY="$work/other-env.json" run_demo ENV=demo; then r=no; else r=yes; fi
check "existing resources count toward the budget" "$r"
check "an existing server is priced" "$(yes_if grep -Eq 'existing server openfief-other \(cax11, nbg1\) +7\.7231' "$work/out")"
check "an unassigned primary IPv4 is priced" "$(yes_if grep -Eq 'existing primary IPv4 spare \(nbg1\) +0\.5950' "$work/out")"
check "the sum includes everything" "$(yes_if grep -Eq 'total +17\.2312 EUR' "$work/out")"
if ACTION=plan INVENTORY="$work/other-env.json" run_demo ENV=demo DEMO_BUDGET_EUR=20; then r=yes; else r=no; fi
check "a larger DEMO_BUDGET_EUR allows it" "$r"

# up from scratch: firewall, then server, with labels, keys, the firewall and the rendered cloud-init.
if ACTION=up run_demo ENV=demo DEMO_SSH_CIDRS=203.0.113.7/32,2001:db8::/64; then r=yes; else r=no; fi
check "up creates a missing demo" "$r"
check "firewall is created with the labels" "$(yes_if grep -qx 'firewall create --name openfief-demo --label app=openfief --label env=demo --rules-file .*' "$work/log")"
check "server is created with labels, key and firewall" "$(yes_if grep -qx 'server create --name openfief-demo --type cax11 --image ubuntu-24.04 --location nbg1 --firewall openfief-demo --ssh-key alice --user-data-from-file .* --label app=openfief --label env=demo' "$work/log")"
check "firewall before server" "$(yes_if bash -c "grep -n create '$work/log' | head -n1 | grep -q firewall")"
check "SSH from the given CIDRs, HTTP(S) from anywhere by default" "$(yes_if gojq -e '. == [
  {"direction":"in","protocol":"tcp","port":"22","source_ips":["203.0.113.7/32","2001:db8::/64"],"description":"ssh (DEMO_SSH_CIDRS)"},
  {"direction":"in","protocol":"tcp","port":"80","source_ips":["0.0.0.0/0","::/0"],"description":"http (DEMO_HTTP_CIDRS)"},
  {"direction":"in","protocol":"tcp","port":"443","source_ips":["0.0.0.0/0","::/0"],"description":"https (DEMO_HTTP_CIDRS)"}]' "$work/state/rules.json")"
check "up prints the demo.openfief.dev URL by default" "$(yes_if grep -q 'demo URL: https://demo.openfief.dev (TLS from Let.s Encrypt prod' "$work/out")"
check "up names the A record the domain needs" "$(yes_if grep -q 'DNS: demo.openfief.dev needs an A record to 192.0.2.10, not proxied' "$work/out")"
check "cloud-init asks Let's Encrypt production by default" "$(yes_if grep -q 'DEMO_ACME_SERVER: https://acme-v02.api.letsencrypt.org/directory' "$work/state/user-data.yaml")"
check "cloud-init pins k3s for arm64" "$(yes_if grep -q "k3s_asset='k3s-arm64'" "$work/state/user-data.yaml")"
check "cloud-init carries the arm64 checksum" "$(yes_if grep -q "$(printf 'a%.0s' {1..64})" "$work/state/user-data.yaml")"
check "cloud-init has no placeholder left" "$(yes_if bash -c "! grep -q '@@' '$work/state/user-data.yaml'")"
check "cloud-init is valid YAML" "$(yes_if gojq --yaml-input -e '.write_files | length > 0' "$work/state/user-data.yaml")"
check "k3s keeps its bundled Traefik and servicelb" "$(yes_if gojq --yaml-input -e '.write_files[] | select(.path == "/etc/rancher/k3s/config.yaml") | .content | contains("disable") | not' "$work/state/user-data.yaml")"
check "cloud-init carries the demo's host" "$(yes_if grep -qx "      host='demo.openfief.dev'" "$work/state/user-data.yaml")"
check "the machine writes the host into the demo settings" "$(yes_if grep -q '"  DEMO_HOST: $host"' "$work/state/user-data.yaml")"
check "SSH takes keys only" "$(yes_if gojq --yaml-input -e '.ssh_pwauth == false and ([.write_files[] | .content | contains("PasswordAuthentication no")] | any)' "$work/state/user-data.yaml")"
check "cloud-init holds no token" "$(yes_if bash -c "! grep -qi 'token' '$work/state/user-data.yaml'")"

# up again: everything exists, so nothing is created; rules are refreshed; an attached firewall is left alone.
if ACTION=up INVENTORY="$work/existing.json" run_demo ENV=demo; then r=yes; else r=no; fi
check "up on an existing demo succeeds" "$r"
check "up is idempotent" "$(yes_if bash -c "! grep -q create '$work/log'")"
check "up refreshes the firewall rules" "$(yes_if grep -qx 'firewall replace-rules --rules-file .* openfief-demo' "$work/log")"
check "an attached firewall is not re-applied" "$(yes_if bash -c "! grep -q apply-to-resource '$work/log'")"
check "no SSH rule without DEMO_SSH_CIDRS" "$(yes_if gojq -e '[.[].port] == ["80", "443"]' "$work/state/rules.json")"
if ACTION=up INVENTORY="$work/existing.json" run_demo ENV=demo DEMO_HTTP_CIDRS=none DEMO_SSH_CIDRS=203.0.113.7/32; then r=yes; else r=no; fi
check "up with DEMO_HTTP_CIDRS=none succeeds" "$r"
check "DEMO_HTTP_CIDRS=none closes HTTP(S)" "$(yes_if gojq -e '[.[].port] == ["22"]' "$work/state/rules.json")"
if ACTION=up run_demo ENV=demo DEMO_ACME=staging; then r=yes; else r=no; fi
check "up with DEMO_ACME=staging succeeds" "$r"
check "staging asks the Let's Encrypt staging server" "$(yes_if grep -q 'DEMO_ACME_SERVER: https://acme-staging-v02.api.letsencrypt.org/directory' "$work/state/user-data.yaml")"
if ACTION=up run_demo ENV=demo DEMO_ACME=production; then r=no; else r=yes; fi
check "an unknown DEMO_ACME is refused" "$r"
check "an unknown DEMO_ACME creates nothing" "$(yes_if bash -c "! grep -q create '$work/log'")"
if ACTION=up run_demo ENV=demo DEMO_DOMAIN=none; then r=yes; else r=no; fi
check "up with DEMO_DOMAIN=none succeeds" "$r"
check "DEMO_DOMAIN=none prints the sslip.io URL" "$(yes_if grep -q 'demo URL: https://192-0-2-10.sslip.io (TLS from Let.s Encrypt prod' "$work/out")"
check "DEMO_DOMAIN=none needs no DNS record" "$(yes_if bash -c "! grep -q '^DNS:' '$work/out'")"
check "DEMO_DOMAIN=none makes the machine name itself on sslip.io" "$(yes_if bash -c "grep -qx \"      host='none'\" '$work/state/user-data.yaml' && grep -q 'host=\"\$(printf .%s. \"\$ipv4\" | tr . -).sslip.io\"' '$work/state/user-data.yaml'")"
if ACTION=up run_demo ENV=demo DEMO_DOMAIN=Not_A_Domain; then r=no; else r=yes; fi
check "a malformed DEMO_DOMAIN is refused" "$r"
check "a malformed DEMO_DOMAIN creates nothing" "$(yes_if bash -c "! grep -q create '$work/log'")"
check "the existing server is not priced twice" "$(yes_if bash -c "! grep -q 'primary IPv4 auto-ip' '$work/out'")"

# Without a labelled SSH key, up refuses and explains the label.
if ACTION=up KEYS="$work/no-keys.json" run_demo ENV=demo; then r=no; else r=yes; fi
check "up without an SSH key is refused" "$r"
check "the refusal explains the label" "$(yes_if grep -q 'label openfief-ssh=true' "$work/out")"
check "no SSH key means nothing created" "$(yes_if bash -c "! grep -q create '$work/log'")"
if ACTION=up run_demo ENV=demo DEMO_SSH_CIDRS=0.0.0.0; then r=no; else r=yes; fi
check "a malformed CIDR is refused" "$r"

# down needs the environment's name, and deletes the server before its firewall.
if ACTION=down run_demo ENV=demo; then r=no; else r=yes; fi
check "down without CONFIRM is refused" "$r"
check "unconfirmed down deletes nothing" "$(yes_if bash -c "! grep -q delete '$work/log'")"
if ACTION=down run_demo ENV=demo CONFIRM=other; then r=no; else r=yes; fi
check "down with another name is refused" "$r"
if ACTION=down run_demo ENV=demo CONFIRM=demo; then r=yes; else r=no; fi
check "confirmed down succeeds" "$r"
check "down selects by both labels" "$(yes_if grep -qx 'all list --selector app=openfief,env=demo --output json' "$work/log")"
check "down deletes server, then firewall" "$(yes_if bash -c "grep delete '$work/log' | tr '\n' ';' | grep -qx 'server delete openfief-demo;firewall delete openfief-demo;primary-ip delete auto-ip;'")"

# The token never appears in arguments or output.
for action in plan up down; do
  ACTION=$action run_demo ENV=demo CONFIRM=demo >/dev/null 2>&1 || true
  check "token stays secret ($action)" "$(yes_if bash -c "! grep -q '$token' '$work/log' '$work/out' '$work/state/'* 2>/dev/null")"
done
if ACTION=plan run_demo ENV=Demo_1; then r=no; else r=yes; fi
check "an invalid ENV is refused" "$r"

# Pure helpers: exact micro-euro parsing, rounding up, architecture and rendering checks.
# helper runs one function of demo.sh in a fresh shell; sourced there, demo.sh skips demo_main.
helper() {
  bash -c 'source "$1"; shift; "$@"' helper "$root/scripts/demo/demo.sh" "$@"
}
check "to_micro is exact" "$(yes_if test "$(helper demo_to_micro 7.7231000000)" = 7723100)"
check "to_micro rounds a seventh digit up" "$(yes_if test "$(helper demo_to_micro 0.0000001)" = 1)"
check "to_micro accepts whole euros" "$(yes_if test "$(helper demo_to_micro 10)" = 10000000)"
check "to_micro reads leading zeros as decimal" "$(yes_if test "$(helper demo_to_micro 0.0800):$(helper demo_to_micro 08.09):$(helper demo_to_micro 0)" = 80000:8090000:0)"
if helper demo_to_micro 1e3 >/dev/null 2>&1; then r=no; else r=yes; fi
check "to_micro rejects text" "$r"
check "format rounds up to four places" "$(yes_if test "$(helper demo_format_micro 7723101)" = 7.7232)"
check "cx23 is amd64" "$(yes_if test "$(helper demo_arch cx23)" = k3s)"
check "cax11 is arm64" "$(yes_if test "$(helper demo_arch cax11)" = k3s-arm64)"
a64="$(printf 'a%.0s' {1..64})"
b64="$(printf 'b%.0s' {1..64})"
pins=(K3S_VERSION=v1.36.5+k3s1 K3S_SHA256_ARM64="$a64" K3S_SHA256_AMD64="$b64"
  K3S_INSTALL_SHA256="$(printf 'c%.0s' {1..64})" FLUX_VERSION=v2.9.6 FLUX_INSTALL_SHA256="$(printf 'd%.0s' {1..64})"
  DEMO_GIT_URL=https://github.com/example/openfief DEMO_GIT_BRANCH=main ENV=demo)

# render renders a template for a server type with the pins above, overridden by NAME=value arguments.
render() {
  local template="$1" type="$2"
  shift 2
  env "${pins[@]}" "$@" bash -c 'source "$1"; demo_render_cloud_init "$2" "$3"' render \
    "$root/scripts/demo/demo.sh" "$template" "$type" 2>/dev/null
}

template="$root/deploy/demo/cloud-init.yaml"
# Output is captured before grep: with pipefail, grep -q closing the pipe early would fail render by SIGPIPE.
amd64="$(render "$template" cx23 || true)"
arm64="$(render "$template" cax11 || true)"
if grep -q "$b64" <<<"$amd64"; then r=yes; else r=no; fi
check "amd64 render uses the amd64 checksum" "$r"
if grep -q 'flux2/releases/download/v2.9.6/install.yaml' <<<"$arm64"; then r=yes; else r=no; fi
check "render pins the Flux version" "$r"
if render "$template" cax11 K3S_SHA256_ARM64= >/dev/null; then r=no; else r=yes; fi
check "render refuses a missing checksum" "$r"
if render "$template" cax11 FLUX_VERSION=latest >/dev/null; then r=no; else r=yes; fi
check "render refuses an unpinned version" "$r"
if render "$template" cax11 $'DEMO_GIT_URL=https://x\ny: z' >/dev/null; then r=no; else r=yes; fi
check "render refuses a git URL with YAML in it" "$r"
printf '#cloud-config\nx: @@UNKNOWN@@\n' >"$work/bad-template.yaml"
if render "$work/bad-template.yaml" cax11 >/dev/null; then r=no; else r=yes; fi
check "render refuses an unknown placeholder" "$r"
custom="$(render "$template" cax11 ENV=try-beta DEMO_DOMAIN=example.org || true)"
if grep -qx "      host='try-beta.example.org'" <<<"$custom"; then r=yes; else r=no; fi
check "render names the host <ENV>.<DEMO_DOMAIN>" "$r"

if [ "$failures" -ne 0 ]; then
  echo "demo: $failures of $cases checks failed" >&2
  exit 1
fi
echo "demo: $cases checks passed"
