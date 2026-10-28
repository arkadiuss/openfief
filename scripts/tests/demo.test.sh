#!/usr/bin/env bash
# Tests for scripts/demo/demo.sh against a fake hcloud with synthetic prices and inventories, and a fake curl
# that plays the Cloudflare DNS API on a record store in a file; no Hetzner or Cloudflare API call
# (kibi: TEST-tooling-demo). Needs gojq and openssl on PATH (task test adds .bin).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
token='fake-token-for-tests-only'
cf_token='fake-cloudflare-token-for-tests-only'
zone_id="$(printf 'z%.0s' {1..32} | tr z f)"

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
  "ssh-key list --selector app=openfief,env="*",openfief-ssh=throwaway --output json") cat "$FAKE_THROWAWAY" ;;
  "firewall describe "*) echo '{"id": 7}' ;;
  "server describe "*) cat "$FAKE_SERVER" ;;
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
  "ssh-key create "*)
    while [ $# -gt 0 ]; do
      if [ "$1" = --public-key-from-file ]; then cp "$2" "$FAKE_STATE/throwaway.pub"; fi
      shift
    done ;;
  *" delete "* | "firewall apply-to-resource "*) ;;
  *) echo "fake hcloud: unexpected call: $*" >&2; exit 2 ;;
esac
EOF
chmod +x "$work/bin/hcloud"

# The fake curl plays the Cloudflare API: it knows the zone example.com (id FAKE_ZONE_ID) and keeps DNS records in
# the JSON array FAKE_CF_STORE. It logs "curl METHOD PATH" and any body, and answers 403-style errors unless the
# Authorization header, read from stdin, carries FAKE_CF_TOKEN. Listing ignores the query's filters on purpose,
# so the script's own filter by comment is what keeps other records safe.
cat >"$work/bin/curl" <<'EOF'
#!/usr/bin/env bash
method=GET url='' body=''
while [ $# -gt 0 ]; do
  case "$1" in
    -X) method="$2"; shift 2 ;;
    --data-binary) body="${2#@}"; shift 2 ;;
    -H | --max-time) [ "$2" != @- ] || header="$(cat)"; shift 2 ;;
    -sS) shift ;;
    https://api.cloudflare.com/client/v4/*) url="$1"; shift ;;
    *) echo "fake curl: unexpected argument: $1" >&2; exit 2 ;;
  esac
done
path="${url#https://api.cloudflare.com/client/v4}"
echo "curl $method $path" >>"$FAKE_LOG"
if [ -n "$body" ]; then echo "body $(gojq -c . "$body")" >>"$FAKE_LOG"; fi
if [ "${header:-}" != "Authorization: Bearer $FAKE_CF_TOKEN" ]; then
  echo '{"success": false, "errors": [{"code": 10000, "message": "Authentication error"}]}'
  exit 0
fi
store="$FAKE_CF_STORE"
ok() { gojq -c '{success: true, result: .}'; }
case "$method $path" in
  "GET /zones?name="*)
    gojq -n -c --arg n "${path#*name=}" --arg id "$FAKE_ZONE_ID" \
      '{success: true, result: [if $n == "example.com" then {id: $id, name: $n} else empty end]}' ;;
  "GET /zones/$FAKE_ZONE_ID/dns_records?"*)
    gojq -c '{success: true, result: ., result_info: {total_count: length}}' "$store" ;;
  "POST /zones/$FAKE_ZONE_ID/dns_records")
    id="$(printf '%032x' "$(($(gojq length "$store") + 1000))")"
    gojq -c --slurpfile r "$body" --arg id "$id" '. + [$r[0] + {id: $id}]' "$store" >"$store.new"
    mv "$store.new" "$store"
    gojq -c --arg id "$id" '.[] | select(.id == $id)' "$store" | ok ;;
  "PATCH /zones/$FAKE_ZONE_ID/dns_records/"*)
    gojq -c --slurpfile r "$body" --arg id "${path##*/}" 'map(if .id == $id then . + $r[0] else . end)' "$store" >"$store.new"
    mv "$store.new" "$store"
    gojq -c --arg id "${path##*/}" '.[] | select(.id == $id)' "$store" | ok ;;
  "DELETE /zones/$FAKE_ZONE_ID/dns_records/"*)
    gojq -c --arg id "${path##*/}" 'map(select(.id != $id))' "$store" >"$store.new"
    mv "$store.new" "$store"
    gojq -n -c --arg id "${path##*/}" '{id: $id}' | ok ;;
  *) echo "fake curl: unexpected call: $method $path" >&2; exit 2 ;;
esac
EOF
chmod +x "$work/bin/curl"

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
echo '[{"id": 3, "name": "openfief-demo-throwaway", "labels": {"app": "openfief", "env": "demo", "openfief-ssh": "throwaway"}}]' \
  >"$work/throwaway.json"
cat >"$work/server.json" <<'EOF'
{"id": 41, "name": "openfief-demo", "labels": {"app": "openfief", "env": "demo"},
 "public_net": {"ipv4": {"ip": "192.0.2.10"}, "ipv6": {"ip": "2001:db8:1:2::/64"}}}
EOF

# DNS records: ours (the comment of env=demo), another demo's, and records the demo doesn't manage.
ours='managed-by=openfief-demo env=demo'
record() { # type name content-or-CAA-value comment id
  gojq -n -c --arg t "$1" --arg n "$2" --arg v "$3" --arg c "$4" --arg id "$(printf '%032x' "$5")" '
    {id: $id, type: $t, name: $n, ttl: 60, comment: (if $c == "" then null else $c end)}
    + if $t == "CAA" then {data: {flags: 0, tag: "issue", value: $v}} else {content: $v, proxied: false} end'
}
echo '[]' >"$work/dns-empty.json"
{
  record A demo.example.com 192.0.2.9 "$ours" 1
  record CAA demo.example.com letsencrypt.org "$ours" 2
  record AAAA demo.example.com 2001:db8:9::1 "$ours" 3
  record A other.demo.example.com 192.0.2.11 'managed-by=openfief-demo env=other' 4
  record A www.example.com 198.51.100.7 '' 5
  record TXT demo.example.com 'v=spf1 -all' '' 6
  record CNAME docs.example.com example.github.io '' 8
} | gojq -s -c . >"$work/dns-mixed.json"
{
  record A demo.example.com 198.51.100.7 '' 7
} | gojq -s -c . >"$work/dns-clash.json"

# run_demo runs demo.sh with the fakes and fresh state; output lands in $work/out, calls in $work/log. Cloudflare's
# records start from DNS (a fixture copied to $work/state/dns.json); CF_TOKEN sets CLOUDFLARE_API_TOKEN.
run_demo() {
  rm -rf "$work/state" "$work/log"
  mkdir -p "$work/state"
  : >"$work/log"
  cp "${DNS:-$work/dns-empty.json}" "$work/state/dns.json"
  local cf=()
  if [ -n "${CF_TOKEN:-}" ]; then cf=(CLOUDFLARE_API_TOKEN="$CF_TOKEN" DEMO_DOMAIN=example.com); fi
  env -u CLOUDFLARE_API_TOKEN -u HETZNER_API_TOKEN -u DEMO_SSH -u DEMO_DOMAIN -u DEMO_DNS_AAAA -u DEMO_SSH_CIDRS -u DOCS_PAGES_HOST \
    PATH="$work/bin:$PATH" HCLOUD_TOKEN="$token" FAKE_LOG="$work/log" FAKE_STATE="$work/state" \
    FAKE_CF_TOKEN="$cf_token" FAKE_CF_STORE="$work/state/dns.json" FAKE_ZONE_ID="$zone_id" \
    FAKE_SERVER="${SERVER:-$work/server.json}" FAKE_THROWAWAY="${THROWAWAY:-$work/no-keys.json}" \
    FAKE_PRICING="$work/pricing.json" FAKE_INVENTORY="${INVENTORY:-$work/empty.json}" \
    FAKE_ENV_INVENTORY="${ENV_INVENTORY:-$work/existing.json}" FAKE_KEYS="${KEYS:-$work/keys.json}" \
    "${cf[@]}" \
    K3S_VERSION=v1.36.5+k3s1 K3S_SHA256_ARM64="$(printf 'a%.0s' {1..64})" K3S_SHA256_AMD64="$(printf 'b%.0s' {1..64})" \
    K3S_INSTALL_SHA256="$(printf 'c%.0s' {1..64})" FLUX_VERSION=v2.9.6 FLUX_INSTALL_SHA256="$(printf 'd%.0s' {1..64})" \
    "$@" bash "$root/scripts/demo/demo.sh" ${MODE:+"$MODE"} "${ACTION:-plan}" >"$work/out" 2>&1
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
check "without a Cloudflare token up names the A record the domain needs" "$(yes_if grep -q 'DNS: demo.openfief.dev needs an A record to 192.0.2.10, not proxied' "$work/out")"
check "cloud-init asks Let's Encrypt production by default" "$(yes_if grep -q 'DEMO_ACME_SERVER: https://acme-v02.api.letsencrypt.org/directory' "$work/state/user-data.yaml")"
check "cloud-init pins k3s for arm64" "$(yes_if grep -q "k3s_asset='k3s-arm64'" "$work/state/user-data.yaml")"
check "cloud-init carries the arm64 checksum" "$(yes_if grep -q "$(printf 'a%.0s' {1..64})" "$work/state/user-data.yaml")"
check "cloud-init has no placeholder left" "$(yes_if bash -c "! grep -q '@@' '$work/state/user-data.yaml'")"
check "cloud-init is valid YAML" "$(yes_if gojq --yaml-input -e '.write_files | length > 0' "$work/state/user-data.yaml")"
check "k3s keeps its bundled Traefik and servicelb" "$(yes_if gojq --yaml-input -e '.write_files[] | select(.path == "/etc/rancher/k3s/config.yaml") | .content | contains("disable") | not' "$work/state/user-data.yaml")"
check "cloud-init carries the demo's host" "$(yes_if grep -qx "      host='demo.openfief.dev'" "$work/state/user-data.yaml")"
check "the machine writes the host into the demo settings" "$(yes_if grep -q '"  DEMO_HOST: $host"' "$work/state/user-data.yaml")"
check "without a Cloudflare token up says so and calls no DNS API" "$(yes_if bash -c "grep -q 'CLOUDFLARE_API_TOKEN is not set' '$work/out' && ! grep -q '^curl' '$work/log'")"
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
if ACTION=up CF_TOKEN="$cf_token" run_demo ENV=demo DEMO_DOMAIN=none; then r=yes; else r=no; fi
check "up with DEMO_DOMAIN=none succeeds" "$r"
check "DEMO_DOMAIN=none prints the sslip.io URL" "$(yes_if grep -q 'demo URL: https://192-0-2-10.sslip.io (TLS from Let.s Encrypt prod' "$work/out")"
check "DEMO_DOMAIN=none skips DNS, even with a Cloudflare token" "$(yes_if bash -c "! grep -q '^DNS' '$work/out' && ! grep -q '^curl' '$work/log'")"
check "DEMO_DOMAIN=none makes the machine name itself on sslip.io" "$(yes_if bash -c "grep -qx \"      host='none'\" '$work/state/user-data.yaml' && grep -q 'host=\"\$(printf .%s. \"\$ipv4\" | tr . -).sslip.io\"' '$work/state/user-data.yaml'")"
if ACTION=up run_demo ENV=demo DEMO_DOMAIN=Not_A_Domain; then r=no; else r=yes; fi
check "a malformed DEMO_DOMAIN is refused" "$r"
check "a malformed DEMO_DOMAIN creates nothing" "$(yes_if bash -c "! grep -q create '$work/log'")"
check "the existing server is not priced twice" "$(yes_if bash -c "! grep -q 'primary IPv4 auto-ip' '$work/out'")"

# Without a labelled SSH key but with SSH sources allowed, up refuses and explains the label and DEMO_SSH=off.
if ACTION=up KEYS="$work/no-keys.json" run_demo ENV=demo DEMO_SSH_CIDRS=203.0.113.7/32; then r=no; else r=yes; fi
check "up without an SSH key but with DEMO_SSH_CIDRS is refused" "$r"
check "the refusal explains the label" "$(yes_if grep -q 'label openfief-ssh=true' "$work/out")"
check "the refusal offers DEMO_SSH=off" "$(yes_if grep -q 'DEMO_SSH=off' "$work/out")"
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

# HETZNER_API_TOKEN is another name for the Hetzner token.
if ACTION=plan run_demo ENV=demo HCLOUD_TOKEN= HETZNER_API_TOKEN="$token"; then r=yes; else r=no; fi
check "HETZNER_API_TOKEN stands in for HCLOUD_TOKEN" "$r"

# DNS with a Cloudflare token: dns_store prints the fake zone's records after a run.
dns_store() {
  gojq -r '[.[] | "\(.type) \(.name) \(.content // .data.value) [\(.comment // "")]"] | sort | join("; ")' \
    "$work/state/dns.json"
}
dns_calls() {
  grep -E '^curl (POST|PATCH|DELETE)' "$work/log" || true
}

# plan shows what DNS would change, and changes nothing.
if ACTION=plan CF_TOKEN="$cf_token" run_demo ENV=demo; then r=yes; else r=no; fi
check "plan with DNS succeeds" "$r"
check "plan looks the zone up by name" "$(yes_if grep -qx 'curl GET /zones?name=example.com' "$work/log")"
check "plan shows the A record for the new server" "$(yes_if grep -qx 'would DNS: create A demo.example.com <IPv4 of the new server>' "$work/out")"
check "plan shows the CAA record" "$(yes_if grep -qx 'would DNS: create CAA demo.example.com 0 issue "letsencrypt.org"' "$work/out")"
check "plan changes no DNS record" "$(yes_if test -z "$(dns_calls)")"

# up creates the server, then the A and CAA records with the demo's comment, DNS-only.
if ACTION=up CF_TOKEN="$cf_token" run_demo ENV=demo; then r=yes; else r=no; fi
check "up with DNS succeeds" "$r"
check "up creates the A and CAA records with the comment" "$(yes_if test "$(dns_store)" = 'A demo.example.com 192.0.2.10 [managed-by=openfief-demo env=demo]; CAA demo.example.com letsencrypt.org [managed-by=openfief-demo env=demo]')"
check "the A record is DNS-only with a short TTL" "$(yes_if gojq -e '.[] | select(.type == "A") | .proxied == false and .ttl == 60' "$work/state/dns.json")"
check "the CAA record allows Let's Encrypt to issue" "$(yes_if gojq -e '.[] | select(.type == "CAA") | .data == {flags: 0, tag: "issue", value: "letsencrypt.org"}' "$work/state/dns.json")"
check "no AAAA record while k3s is IPv4-only" "$(yes_if bash -c "! grep -q AAAA '$work/state/dns.json'")"
check "DNS records come after the server" "$(yes_if bash -c "grep -nE '^(server create|curl POST)' '$work/log' | head -n1 | grep -q 'server create'")"
check "cloud-init carries the DNS host" "$(yes_if grep -qF "host='demo.example.com'" "$work/state/user-data.yaml")"
check "up prints the DNS URL" "$(yes_if grep -q 'demo URL: https://demo.example.com (TLS' "$work/out")"

# A second up changes nothing; an up after a new IPv4 updates the A record in place.
cp "$work/state/dns.json" "$work/dns-ours.json"
if ACTION=up CF_TOKEN="$cf_token" DNS="$work/dns-ours.json" INVENTORY="$work/existing.json" run_demo ENV=demo; then r=yes; else r=no; fi
check "up again with DNS succeeds" "$r"
check "up again changes no DNS record" "$(yes_if test -z "$(dns_calls)")"
check "up again says DNS is up to date" "$(yes_if grep -q 'DNS records of demo.example.com are up to date' "$work/out")"
sed 's/192\.0\.2\.10/192.0.2.20/' "$work/server.json" >"$work/server-new-ip.json"
if ACTION=up CF_TOKEN="$cf_token" DNS="$work/dns-ours.json" SERVER="$work/server-new-ip.json" run_demo ENV=demo; then r=yes; else r=no; fi
check "up after an IP change succeeds" "$r"
check "the A record is updated, not added" "$(yes_if test "$(dns_calls | cut -d' ' -f2 | tr '\n' ' ')" = 'PATCH ')"
check "the A record points at the new IPv4" "$(yes_if gojq -e '[.[] | select(.type == "A")] | length == 1 and .[0].content == "192.0.2.20"' "$work/state/dns.json")"
check "the update is reported" "$(yes_if grep -q 'DNS: update A demo.example.com 192.0.2.10 -> 192.0.2.20' "$work/out")"

# With other records in the zone: ours are reconciled, the others are left alone.
if ACTION=up CF_TOKEN="$cf_token" DNS="$work/dns-mixed.json" run_demo ENV=demo; then r=yes; else r=no; fi
check "up in a shared zone succeeds" "$r"
check "only our records change" "$(yes_if test "$(dns_store)" = 'A demo.example.com 192.0.2.10 [managed-by=openfief-demo env=demo]; A other.demo.example.com 192.0.2.11 [managed-by=openfief-demo env=other]; A www.example.com 198.51.100.7 []; CAA demo.example.com letsencrypt.org [managed-by=openfief-demo env=demo]; CNAME docs.example.com example.github.io []; TXT demo.example.com v=spf1 -all []')"
check "our stale AAAA record is deleted" "$(yes_if grep -q 'DNS: delete AAAA demo.example.com 2001:db8:9::1' "$work/out")"
if ACTION=up CF_TOKEN="$cf_token" DNS="$work/dns-mixed.json" run_demo ENV=demo DEMO_DNS_AAAA=on; then r=yes; else r=no; fi
check "DEMO_DNS_AAAA=on points AAAA at the server's IPv6" "$(yes_if gojq -e '[.[] | select(.type == "AAAA")] | length == 1 and .[0].content == "2001:db8:1:2::1"' "$work/state/dns.json")"
if ACTION=up CF_TOKEN="$cf_token" DNS="$work/dns-clash.json" run_demo ENV=demo; then r=no; else r=yes; fi
check "a record set by hand at the host is refused" "$r"
check "the clash is named" "$(yes_if grep -q 'refusing: the zone has A demo.example.com without the comment' "$work/out")"
check "a clash creates nothing" "$(yes_if bash -c "! grep -qE '^(server|firewall) create' '$work/log' && test -z '$(dns_calls)'")"
check "a clash leaves the record without a comment as it was" "$(yes_if test "$(dns_store)" = 'A demo.example.com 198.51.100.7 []')"
if ACTION=down CF_TOKEN="$cf_token" DNS="$work/dns-clash.json" run_demo ENV=demo CONFIRM=demo; then r=yes; else r=no; fi
check "down beside a record without a comment succeeds" "$r"
check "down never deletes a record without the comment" "$(yes_if bash -c "test -z '$(dns_calls)' && test \"\$(gojq length '$work/state/dns.json')\" = 1")"
if ACTION=up CF_TOKEN="$cf_token" run_demo ENV=try-beta ENV_INVENTORY="$work/empty.json"; then r=yes; else r=no; fi
check "another demo lives at <env>.<domain>" "$(yes_if gojq -e 'map(.name) | unique == ["try-beta.example.com"]' "$work/state/dns.json")"

# down deletes exactly the records with this demo's comment; recreate keeps them and points them at the new server.
if ACTION=down CF_TOKEN="$cf_token" DNS="$work/dns-mixed.json" run_demo ENV=demo CONFIRM=demo; then r=yes; else r=no; fi
check "down with DNS succeeds" "$r"
check "down deletes only our records" "$(yes_if test "$(dns_store)" = 'A other.demo.example.com 192.0.2.11 [managed-by=openfief-demo env=other]; A www.example.com 198.51.100.7 []; CNAME docs.example.com example.github.io []; TXT demo.example.com v=spf1 -all []')"
check "down deletes DNS after the server" "$(yes_if bash -c "grep -nE '^(server delete|curl DELETE)' '$work/log' | head -n1 | grep -q 'server delete'")"
if ACTION=recreate CF_TOKEN="$cf_token" DNS="$work/dns-ours.json" SERVER="$work/server-new-ip.json" run_demo ENV=demo CONFIRM=demo; then r=yes; else r=no; fi
check "recreate with DNS succeeds" "$r"
check "recreate keeps the records and updates the A record" "$(yes_if test "$(dns_calls | cut -d' ' -f2 | tr '\n' ' ')" = 'PATCH ')"
if ACTION=down run_demo ENV=demo CONFIRM=demo; then r=yes; else r=no; fi
check "down without a Cloudflare token says the records stay" "$(yes_if grep -q 'DNS records of demo .demo. stay' "$work/out")"

# status lists our records; a token Cloudflare rejects, or one without the zone, fails before anything is created.
if ACTION=status CF_TOKEN="$cf_token" DNS="$work/dns-mixed.json" INVENTORY="$work/existing.json" run_demo ENV=demo; then r=yes; else r=no; fi
check "status shows our records only" "$(yes_if bash -c "grep -c '^DNS record ' '$work/out' | grep -qx 3 && grep -qx 'DNS record A demo.example.com 192.0.2.9' '$work/out'")"
if ACTION=up CF_TOKEN=wrong-token run_demo ENV=demo; then r=no; else r=yes; fi
check "a rejected Cloudflare token fails up" "$r"
check "the API's error is shown" "$(yes_if grep -q 'Cloudflare API: GET failed: 10000 Authentication error' "$work/out")"
check "a rejected token creates nothing" "$(yes_if bash -c "! grep -q create '$work/log'")"
if ACTION=plan CF_TOKEN="$cf_token" run_demo ENV=demo DEMO_DOMAIN=example.org; then r=no; else r=yes; fi
check "a zone the token can't see is refused" "$r"
check "the refusal names the permissions" "$(yes_if grep -q 'Zone > DNS > Edit' "$work/out")"
for action in plan up down; do
  ACTION=$action CF_TOKEN="$cf_token" DNS="$work/dns-mixed.json" run_demo ENV=demo CONFIRM=demo >/dev/null 2>&1 || true
  check "Cloudflare token stays secret ($action)" "$(yes_if bash -c "! grep -q '$cf_token' '$work/log' '$work/out' '$work/state/'* 2>/dev/null")"
done

# SSH off: with no labelled key and no SSH source, up installs a throwaway public key and nothing else.
if ACTION=up KEYS="$work/no-keys.json" run_demo ENV=demo; then r=yes; else r=no; fi
check "up without keys or SSH sources succeeds" "$r"
check "the throwaway key is uploaded with its labels" "$(yes_if grep -qx 'ssh-key create --name openfief-demo-throwaway --public-key-from-file .* --label app=openfief --label env=demo --label openfief-ssh=throwaway' "$work/log")"
check "the server gets only the throwaway key" "$(yes_if grep -qx 'server create .* --firewall openfief-demo --ssh-key openfief-demo-throwaway --user-data-from-file .*' "$work/log")"
check "only an Ed25519 public key is uploaded" "$(yes_if grep -Eqx 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI[A-Za-z0-9+/]{43} openfief-demo-throwaway' "$work/state/throwaway.pub")"
check "up says SSH is off" "$(yes_if grep -q 'SSH is off' "$work/out")"
check "the firewall has no SSH rule" "$(yes_if gojq -e '[.[].port] == ["80", "443"]' "$work/state/rules.json")"
if ACTION=up KEYS="$work/no-keys.json" THROWAWAY="$work/throwaway.json" run_demo ENV=demo; then r=yes; else r=no; fi
check "an existing throwaway key is reused" "$(yes_if bash -c "! grep -q 'ssh-key create' '$work/log' && grep -q -- '--ssh-key openfief-demo-throwaway' '$work/log'")"
if ACTION=up run_demo ENV=demo DEMO_SSH=off; then r=yes; else r=no; fi
check "DEMO_SSH=off ignores labelled keys" "$(yes_if bash -c "grep -q -- '--ssh-key openfief-demo-throwaway' '$work/log' && ! grep -q -- '--ssh-key alice' '$work/log'")"
if ACTION=plan KEYS="$work/no-keys.json" run_demo ENV=demo; then r=yes; else r=no; fi
check "plan says SSH would be off" "$(yes_if grep -q 'SSH off: a new server gets a throwaway key' "$work/out")"
if ACTION=up run_demo ENV=demo DEMO_SSH=off DEMO_SSH_CIDRS=203.0.113.7/32; then r=no; else r=yes; fi
check "DEMO_SSH=off with SSH sources is refused" "$r"
if ACTION=up KEYS="$work/no-keys.json" run_demo ENV=demo DEMO_SSH=keys; then r=no; else r=yes; fi
check "DEMO_SSH=keys without a key is refused" "$r"
if ACTION=up run_demo ENV=demo DEMO_SSH=maybe; then r=no; else r=yes; fi
check "an unknown DEMO_SSH is refused" "$r"
check "a refused SSH setting creates nothing" "$(yes_if bash -c "! grep -q create '$work/log'")"
if ACTION=down THROWAWAY="$work/throwaway.json" run_demo ENV=demo CONFIRM=demo; then r=yes; else r=no; fi
check "down deletes the throwaway key after the server" "$(yes_if bash -c "grep delete '$work/log' | tr '\n' ';' | grep -qx 'server delete openfief-demo;firewall delete openfief-demo;primary-ip delete auto-ip;ssh-key delete openfief-demo-throwaway;'")"
check "down asks only for throwaway keys" "$(yes_if grep -qx 'ssh-key list --selector app=openfief,env=demo,openfief-ssh=throwaway --output json' "$work/log")"

# The docs site's record (demo.sh dns-docs, task dns:docs): docs.<domain> as a DNS-only CNAME to GitHub Pages with its
# own comment; it never touches the demo's records, and the demo never touches it.
docs_comment='managed-by=openfief-docs'
{
  record CNAME docs.example.com arkadiuss.github.io "$docs_comment" 20
} | gojq -s -c . >"$work/dns-docs.json"
{
  record CNAME docs.example.com someone.github.io "$docs_comment" 20
} | gojq -s -c . >"$work/dns-docs-old.json"
{
  record TXT docs.example.com 'verification' '' 21
} | gojq -s -c . >"$work/dns-docs-foreign.json"
gojq -c '. + [{id: "00000000000000000000000000000014", type: "CNAME", name: "docs.example.com",
  content: "arkadiuss.github.io", ttl: 60, proxied: false, comment: "managed-by=openfief-docs"}]' \
  "$work/dns-mixed.json" >"$work/dns-mixed-docs.json"
if ACTION=plan MODE=dns-docs CF_TOKEN="$cf_token" run_demo; then r=yes; else r=no; fi
check "dns:docs:plan succeeds" "$r"
check "dns:docs:plan shows the CNAME" "$(yes_if grep -qx 'would DNS: create CNAME docs.example.com arkadiuss.github.io' "$work/out")"
check "dns:docs:plan changes nothing" "$(yes_if test -z "$(dns_calls)")"
if ACTION=apply MODE=dns-docs CF_TOKEN="$cf_token" run_demo; then r=yes; else r=no; fi
check "dns:docs creates the CNAME with its comment, DNS-only" "$(yes_if gojq -e '. == [.[0]] and .[0].type == "CNAME" and .[0].name == "docs.example.com" and .[0].content == "arkadiuss.github.io" and .[0].proxied == false and .[0].comment == "managed-by=openfief-docs"' "$work/state/dns.json")"
if ACTION=apply MODE=dns-docs CF_TOKEN="$cf_token" DNS="$work/dns-docs.json" run_demo; then r=yes; else r=no; fi
check "dns:docs again changes nothing" "$(yes_if bash -c "test -z '$(dns_calls)' && grep -q 'docs.example.com is up to date' '$work/out'")"
if ACTION=apply MODE=dns-docs CF_TOKEN="$cf_token" DNS="$work/dns-docs-old.json" run_demo; then r=yes; else r=no; fi
check "dns:docs updates its own record in place" "$(yes_if bash -c "test \"\$(grep -cE '^curl (POST|PATCH|DELETE)' '$work/log')\" = 1 && grep -q '^curl PATCH' '$work/log' && grep -q 'arkadiuss.github.io' '$work/state/dns.json'")"
if ACTION=apply MODE=dns-docs CF_TOKEN="$cf_token" DNS="$work/dns-docs-foreign.json" run_demo; then r=no; else r=yes; fi
check "dns:docs refuses any record without its comment at the name" "$r"
check "dns:docs leaves that record alone" "$(yes_if bash -c "test -z '$(dns_calls)' && cmp -s '$work/dns-docs-foreign.json' '$work/state/dns.json'")"
gojq -c 'map(select(.name != "docs.example.com"))' "$work/dns-mixed.json" >"$work/dns-mixed-nodocs.json"
if ACTION=apply MODE=dns-docs CF_TOKEN="$cf_token" DNS="$work/dns-mixed-nodocs.json" run_demo; then r=yes; else r=no; fi
check "dns:docs never touches the demo's records" "$(yes_if bash -c "test \"\$(grep -cE '^curl (POST|PATCH|DELETE)' '$work/log')\" = 1 && grep -q '^curl POST' '$work/log'")"
if ACTION=down CF_TOKEN="$cf_token" DNS="$work/dns-mixed-docs.json" run_demo ENV=demo CONFIRM=demo; then r=yes; else r=no; fi
check "the demo's down keeps the docs record" "$(yes_if gojq -e '[.[] | select(.comment == "managed-by=openfief-docs")] | length == 1' "$work/state/dns.json")"
if ACTION=up CF_TOKEN="$cf_token" DNS="$work/dns-docs.json" run_demo ENV=docs ENV_INVENTORY="$work/empty.json"; then r=no; else r=yes; fi
check "a demo named docs is refused at the docs record's name" "$r"
check "the refused demo leaves the docs record alone" "$(yes_if cmp -s "$work/dns-docs.json" "$work/state/dns.json")"
if ACTION=plan MODE=dns-docs run_demo; then r=no; else r=yes; fi
check "dns:docs needs CLOUDFLARE_API_TOKEN" "$r"
check "dns:docs without a token names it and calls nothing" "$(yes_if bash -c "grep -q 'CLOUDFLARE_API_TOKEN is not set' '$work/out' && ! grep -q '^curl' '$work/log'")"
if ACTION=plan MODE=dns-docs CF_TOKEN="$cf_token" run_demo DEMO_DOMAIN=none; then r=no; else r=yes; fi
check "dns:docs refuses DEMO_DOMAIN=none" "$r"
for action in plan apply; do
  ACTION=$action MODE=dns-docs CF_TOKEN="$cf_token" DNS="$work/dns-docs-old.json" run_demo >/dev/null 2>&1 || true
  check "Cloudflare token stays secret (dns:docs $action)" "$(yes_if bash -c "! grep -q '$cf_token' '$work/log' '$work/out' '$work/state/'* 2>/dev/null")"
done

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
  DEMO_GIT_URL=https://github.com/example/openfief DEMO_GIT_BRANCH=main)

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
if grep -qF "host='none'" <<<"$arm64"; then r=yes; else r=no; fi
check "render defaults to the machine's sslip.io host" "$r"
if env "${pins[@]}" bash -c 'source "$1"; demo_render_cloud_init "$2" cax11 "demo.openfief.dev"' render \
  "$root/scripts/demo/demo.sh" "$template" 2>/dev/null | grep -qF "host='demo.openfief.dev'"; then r=yes; else r=no; fi
check "render writes a DNS host" "$r"
if env "${pins[@]}" bash -c 'source "$1"; demo_render_cloud_init "$2" cax11 "x; rm -rf /"' render \
  "$root/scripts/demo/demo.sh" "$template" >/dev/null 2>&1; then r=no; else r=yes; fi
check "render refuses a host that isn't a DNS name" "$r"

# Host names, IPv6 addresses and the pure DNS diff.
check "the demo named demo is demo.<domain>" "$(yes_if test "$(helper demo_host demo openfief.dev)" = demo.openfief.dev)"
check "other demos are <env>.<domain>" "$(yes_if test "$(helper demo_host try-beta example.org)" = try-beta.example.org)"
check "DEMO_DOMAIN=none gives no host" "$(yes_if test "$(helper demo_host demo none)" = none)"
if helper demo_host demo 'Example.COM' >/dev/null 2>&1; then r=no; else r=yes; fi
check "demo_host refuses a malformed domain" "$r"
check "a server's IPv6 is ::1 of its network" "$(yes_if test "$(helper demo_ipv6_host 2001:db8:1:2::/64)" = 2001:db8:1:2::1)"
check "no IPv6 network gives no address" "$(yes_if test -z "$(helper demo_ipv6_host '')")"
helper demo_dns_desired demo.example.com 192.0.2.10 '' "$ours" >"$work/desired.json"
check "no changes when the zone matches" "$(yes_if test -z "$(helper demo_dns_changes "$work/desired.json" "$work/dns-ours.json" "$ours" demo.example.com 2>/dev/null || echo failed)")"
gojq -c 'map(.proxied = true)' "$work/dns-ours.json" >"$work/dns-proxied.json"
check "a proxied record is set back to DNS-only" "$(yes_if test "$(helper demo_dns_changes "$work/desired.json" "$work/dns-proxied.json" "$ours" demo.example.com | gojq -r .op | sort -u)" = update)"
gojq -c '. + [.[0] | .id = "0000000000000000000000000000ffff"]' "$work/dns-ours.json" >"$work/dns-dup.json"
check "a duplicate of our record is deleted" "$(yes_if test "$(helper demo_dns_changes "$work/desired.json" "$work/dns-dup.json" "$ours" demo.example.com | gojq -r '"\(.op) \(.id)"')" = 'delete 0000000000000000000000000000ffff')"
der="$work/der-only"
mkdir -p "$der"
helper demo_throwaway_pubkey "$der/key.der" test-key >/dev/null
check "a throwaway key leaves only its public half on disk" "$(yes_if bash -c "test \"\$(ls '$der')\" = key.der && ! grep -qa PRIVATE '$der/key.der'")"

if [ "$failures" -ne 0 ]; then
  echo "demo: $failures of $cases checks failed" >&2
  exit 1
fi
echo "demo: $cases checks passed"
