#!/usr/bin/env bash
# Creates, inspects and destroys a demo machine on Hetzner Cloud (docs/dev/demo.md). Called through task:
#   task demo:plan ENV=demo        what `up` would create and the monthly cost; changes nothing
#   task demo:up ENV=demo          create what is missing (idempotent), within DEMO_BUDGET_EUR
#   task demo:status ENV=demo      the environment's resources and cost
#   task demo:down ENV=demo CONFIRM=demo      delete every resource labelled app=openfief,env=demo
#   task demo:recreate ENV=demo CONFIRM=demo  down, then up
# Stateless: every resource carries the labels app=openfief and env=<ENV> and is found by them.
# The API token comes from HCLOUD_TOKEN only; it is never printed or written to disk.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"

# Pure helpers: nothing below up to demo_prepare calls the Hetzner API. Inputs are JSON documents and
# environment variables, outputs go to stdout, and money is integer micro-euros, never floating point.

# demo_fail prints a message to stderr and fails.
demo_fail() {
  echo "demo: $*" >&2
  return 1
}

# demo_check_env accepts a demo environment name: it becomes a label value and part of resource names.
demo_check_env() {
  if [[ ! "${1:-}" =~ ^[a-z][a-z0-9-]{0,19}$ ]]; then
    demo_fail "ENV '${1:-}' must be 1-20 lowercase letters, digits or dashes, starting with a letter (e.g. ENV=demo)"
    return 1
  fi
}

# demo_to_micro converts a non-negative decimal string such as "7.7231000000" to integer micro-euros,
# rounding any digit beyond the sixth up, so a budget check never under-counts.
demo_to_micro() {
  local value="$1" digits frac rest micro
  if [[ ! "$value" =~ ^([0-9]{1,9})(\.([0-9]*))?$ ]]; then
    demo_fail "not a price: '$value'"
    return 1
  fi
  frac="${BASH_REMATCH[3]}000000"
  rest="${frac:6}"
  # Whole euros and six decimals as one digit string, without leading zeros so bash doesn't read it as octal.
  digits="${BASH_REMATCH[1]}${frac:0:6}"
  while [[ "$digits" == 0?* ]]; do digits="${digits#0}"; done
  micro=$((digits))
  if [[ "$rest" =~ [1-9] ]]; then
    micro=$((micro + 1))
  fi
  echo "$micro"
}

# demo_format_micro prints micro-euros as euros with four decimals, rounded up.
demo_format_micro() {
  local tenth_mils=$((($1 + 99) / 100))
  printf '%d.%04d' $((tenth_mils / 10000)) $((tenth_mils % 10000))
}

# demo_cost_lines prints one tab-separated line per billable item: description, monthly gross price, quantity and
# percentage. Inputs: the Hetzner pricing document (GET /pricing), the inventory of everything labelled
# app=openfief (hcloud all list -o json) and the planned server ({"servers":[{name,type,location}]}).
# A planned server always has a primary IPv4: GitHub and GHCR, which Flux pulls from, have no IPv6.
demo_cost_lines() {
  gojq -r --slurpfile pricing "$1" --slurpfile inventory "$2" --slurpfile planned "$3" -n '
    ($pricing[0].pricing // $pricing[0]) as $p
    | if $p.currency != "EUR" then error("pricing currency is \($p.currency), the budget is in EUR") else . end
    | def server_price($type; $loc):
        ([$p.server_types[] | select(.name == $type) | .prices[] | select(.location == $loc) | .price_monthly.gross]
          | first) // error("server type \($type) has no price at \($loc)");
      def ipv4_price($loc):
        ([$p.primary_ips[] | select(.type == "ipv4") | .prices[] | select(.location == $loc) | .price_monthly.gross]
          | first) // error("primary IPv4 has no price at \($loc)");
      def line($desc; $price; $qty; $pct): [$desc, $price, ($qty | tostring), $pct] | join("\t");
      ($inventory[0] // {}) as $inv
    | ([($inv.servers // [])[] | .id]) as $server_ids
    | (($inv.servers // [])[]
        | . as $s
        | line("existing server \($s.name) (\($s.server_type.name), \($s.location.name))";
               server_price($s.server_type.name; $s.location.name); 1; "100"),
          (if (($s.public_net.ipv4.ip // "") != "") then
             line("existing IPv4 of \($s.name)"; ipv4_price($s.location.name); 1; "100") else empty end),
          (if $s.backup_window != null then
             line("existing backups of \($s.name)"; server_price($s.server_type.name; $s.location.name); 1;
                  $p.server_backup.percentage) else empty end)),
      (($inv.primary_ips // [])[]
        | select(.type == "ipv4")
        | select(.assignee_id == null or (.assignee_id as $a | $server_ids | any(.[]; . == $a) | not))
        | line("existing primary IPv4 \(.name) (\(.location.name))"; ipv4_price(.location.name); 1; "100")),
      (($inv.floating_ips // [])[]
        | . as $f
        | line("existing floating \($f.type) \($f.name) (\($f.home_location.name))";
               ([$p.floating_ips[]? | select(.type == $f.type) | .prices[] | select(.location == $f.home_location.name)
                 | .price_monthly.gross] | first) // error("floating IP has no price at \($f.home_location.name)");
               1; "100")),
      (($inv.volumes // [])[]
        | line("existing volume \(.name) (\(.size) GB)"; $p.volume.price_per_gb_month.gross; .size; "100")),
      (($inv.images // [])[]
        | line("existing snapshot \(.description // .id) (\(.image_size // 0 | ceil) GB)";
               $p.image.price_per_gb_month.gross; (.image_size // 0 | ceil); "100")),
      (($inv.load_balancers // [])[]
        | . as $lb
        | line("existing load balancer \($lb.name) (\($lb.load_balancer_type.name))";
               ([$p.load_balancer_types[] | select(.name == $lb.load_balancer_type.name) | .prices[]
                 | select(.location == $lb.location.name) | .price_monthly.gross] | first)
                 // error("load balancer type has no price at \($lb.location.name)");
               1; "100")),
      (($planned[0].servers // [])[]
        | line("new server \(.name) (\(.type), \(.location))"; server_price(.type; .location); 1; "100"),
          line("new IPv4 of \(.name)"; ipv4_price(.location); 1; "100"))
  '
}

# demo_budget_check reads demo_cost_lines output on stdin, prints the arithmetic, and fails when the monthly
# gross total exceeds the budget in euros (the first argument).
demo_budget_check() {
  local budget_eur="$1" budget total=0 desc price qty pct unit percent amount
  budget="$(demo_to_micro "$budget_eur")" || return 1
  while IFS=$'\t' read -r desc price qty pct; do
    [ -n "$desc" ] || continue
    unit="$(demo_to_micro "$price")" || return 1
    percent="$(demo_to_micro "$pct")" || return 1
    # amount = unit * qty * pct / 100, rounded up; percent is in micro-units, hence 10^8.
    amount=$(((unit * qty * percent + 99999999) / 100000000))
    total=$((total + amount))
    if [ "$qty" = 1 ] && [ "$pct" = 100 ]; then
      printf '  %-58s %10s EUR\n' "$desc" "$(demo_format_micro "$amount")"
    else
      printf '  %-58s %10s EUR (%s x %s x %s%%)\n' "$desc" "$(demo_format_micro "$amount")" "$price" "$qty" "$pct"
    fi
  done
  printf '  %-58s %10s EUR per month, gross\n' "total" "$(demo_format_micro "$total")"
  printf '  %-58s %10s EUR (DEMO_BUDGET_EUR)\n' "budget" "$(demo_format_micro "$budget")"
  if [ "$total" -gt "$budget" ]; then
    demo_fail "refusing: $(demo_format_micro "$total") EUR per month would exceed the budget of $(demo_format_micro "$budget") EUR"
    return 1
  fi
}

# demo_arch maps a Hetzner server type to the k3s release asset for its CPU: CAX types are Arm64.
demo_arch() {
  case "$1" in
    cax*) echo k3s-arm64 ;;
    *) echo k3s ;;
  esac
}

# demo_cidrs prints a space- or comma-separated list of CIDRs (the second argument, from the setting named by
# the first) as a JSON array; "none" or nothing gives []. It fails on anything that isn't an IPv4 or IPv6 CIDR.
demo_cidrs() {
  local setting="$1" cidr list=()
  for cidr in ${2//,/ }; do
    [ "$cidr" != none ] || continue
    if [[ ! "$cidr" =~ ^[0-9.]+/[0-9]{1,2}$ && ! "$cidr" =~ ^[0-9a-fA-F:]+/[0-9]{1,3}$ ]]; then
      demo_fail "$setting: '$cidr' is not an IPv4 or IPv6 CIDR such as 203.0.113.7/32"
      return 1
    fi
    list+=("$cidr")
  done
  if [ "${#list[@]}" -eq 0 ]; then
    echo '[]'
    return 0
  fi
  printf '%s\n' "${list[@]}" | gojq -R . | gojq -s -c .
}

# demo_firewall_rules prints the Hetzner firewall rules: SSH (22/tcp) from the CIDRs in the first argument
# (DEMO_SSH_CIDRS), HTTP and HTTPS (80 and 443/tcp) from those in the second (DEMO_HTTP_CIDRS). A port with no
# CIDRs gets no rule, so with neither the firewall blocks all inbound traffic; outbound stays open.
demo_firewall_rules() {
  local ssh http
  ssh="$(demo_cidrs DEMO_SSH_CIDRS "$1")" || return 1
  http="$(demo_cidrs DEMO_HTTP_CIDRS "$2")" || return 1
  gojq -n -c --argjson ssh "$ssh" --argjson http "$http" '
    def rule($port; $from; $what): if ($from | length) > 0
      then {direction: "in", protocol: "tcp", port: $port, source_ips: $from, description: $what} else empty end;
    [rule("22"; $ssh; "ssh (DEMO_SSH_CIDRS)"), rule("80"; $http; "http (DEMO_HTTP_CIDRS)"),
     rule("443"; $http; "https (DEMO_HTTP_CIDRS)")]'
}

# demo_ssh_keys prints the names of the SSH keys in an `hcloud ssh-key list -o json` document, and fails with
# instructions when there is none.
demo_ssh_keys() {
  local names
  names="$(gojq -r '.[].name' "$1")"
  if [ -z "$names" ]; then
    demo_fail "no SSH key in the project has the label openfief-ssh=true; add one with:
  hcloud ssh-key create --name <you> --public-key-from-file ~/.ssh/id_ed25519.pub --label openfief-ssh=true
or label an existing key: hcloud ssh-key add-label <key> openfief-ssh=true"
    return 1
  fi
  echo "$names"
}

# demo_host is the demo's host name, <ENV>.<DEMO_DOMAIN>, or "none" when DEMO_DOMAIN=none: the machine then names
# itself <ipv4-with-dashes>.sslip.io at boot. demo_render_cloud_init checks its form.
demo_host() {
  if [ "${DEMO_DOMAIN:-}" = none ]; then echo none; else echo "${ENV:-}.${DEMO_DOMAIN:-}"; fi
}

# demo_render_cloud_init fills the @@NAME@@ placeholders of a cloud-init template from the pinned versions in the
# environment, after checking each value's form, and fails if a placeholder is left.
demo_render_cloud_init() {
  local template="$1" server_type="$2" asset k3s_sha acme rendered name value
  asset="$(demo_arch "$server_type")"
  if [ "$asset" = k3s-arm64 ]; then k3s_sha="${K3S_SHA256_ARM64:-}"; else k3s_sha="${K3S_SHA256_AMD64:-}"; fi
  # Let's Encrypt production by default; staging for experiments, whose certificates browsers don't trust but
  # whose rate limits are far higher.
  case "${DEMO_ACME:-prod}" in
    prod) acme=https://acme-v02.api.letsencrypt.org/directory ;;
    staging) acme=https://acme-staging-v02.api.letsencrypt.org/directory ;;
    *) acme="unknown DEMO_ACME '${DEMO_ACME:-}', want prod or staging" ;;
  esac
  local -A values=(
    [ACME_SERVER]="$acme" [DEMO_HOST]="$(demo_host)"
    [K3S_VERSION]="${K3S_VERSION:-}" [K3S_ASSET]="$asset" [K3S_SHA256]="$k3s_sha"
    [K3S_INSTALL_SHA256]="${K3S_INSTALL_SHA256:-}" [FLUX_VERSION]="${FLUX_VERSION:-}"
    [FLUX_INSTALL_SHA256]="${FLUX_INSTALL_SHA256:-}"
    [GIT_URL]="${DEMO_GIT_URL:-}" [GIT_BRANCH]="${DEMO_GIT_BRANCH:-}"
  )
  local -A forms=(
    [ACME_SERVER]='^https://acme(-staging)?-v02\.api\.letsencrypt\.org/directory$'
    [DEMO_HOST]='^(none|([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63})$'
    [K3S_VERSION]='^v[0-9]+\.[0-9]+\.[0-9]+\+k3s[0-9]+$' [K3S_ASSET]='^k3s(-arm64)?$'
    [K3S_SHA256]='^[0-9a-f]{64}$' [K3S_INSTALL_SHA256]='^[0-9a-f]{64}$'
    [FLUX_VERSION]='^v[0-9]+\.[0-9]+\.[0-9]+$' [FLUX_INSTALL_SHA256]='^[0-9a-f]{64}$'
    [GIT_URL]='^https://[A-Za-z0-9./_-]+$' [GIT_BRANCH]='^[A-Za-z0-9._/-]+$'
  )
  rendered="$(cat "$template")"
  for name in "${!values[@]}"; do
    value="${values[$name]}"
    if [[ ! "$value" =~ ${forms[$name]} ]]; then
      demo_fail "cloud-init: $name='$value' does not match ${forms[$name]}"
      return 1
    fi
    rendered="${rendered//@@${name}@@/$value}"
  done
  if [[ "$rendered" =~ @@[A-Z0-9_]+@@ ]]; then
    demo_fail "cloud-init: unfilled placeholder ${BASH_REMATCH[0]} in $template"
    return 1
  fi
  printf '%s\n' "$rendered"
}

: "${DEMO_SERVER_TYPE:=cax11}" "${DEMO_LOCATION:=nbg1}" "${DEMO_IMAGE:=ubuntu-24.04}" "${DEMO_BUDGET_EUR:=10}"
: "${DEMO_SSH_CIDRS:=}" "${DEMO_GIT_URL:=https://github.com/arkadiuss/openfief}" "${DEMO_GIT_BRANCH:=main}"
# HTTP(S) is open to everyone by default: a demo holds synthetic data only (GR-12). DEMO_HTTP_CIDRS=none closes it.
: "${DEMO_HTTP_CIDRS:=0.0.0.0/0,::/0}" "${DEMO_ACME:=prod}"
# The demo ENV is served at <ENV>.<DEMO_DOMAIN>, such as demo.openfief.dev; DEMO_DOMAIN=none uses sslip.io instead.
: "${DEMO_DOMAIN:=openfief.dev}"
export DEMO_GIT_URL DEMO_GIT_BRANCH DEMO_ACME DEMO_DOMAIN

# demo_prepare checks the token and tools and isolates hcloud from any local CLI context.
demo_prepare() {
  if [ -z "${HCLOUD_TOKEN:-}" ]; then
    demo_fail "HCLOUD_TOKEN is not set. Create a Read & Write API token in the Hetzner Cloud console (project >
Security > API tokens) and export it in this shell only, e.g. read -rs HCLOUD_TOKEN && export HCLOUD_TOKEN"
    return 1
  fi
  local tool
  for tool in hcloud gojq; do
    command -v "$tool" >/dev/null 2>&1 || demo_fail "$tool not found; run task setup" || return 1
  done
  # A missing config file: hcloud uses HCLOUD_TOKEN, never a stored context of another project.
  export HCLOUD_CONFIG="$work/no-cli.toml"
  unset HCLOUD_CONTEXT
}

# demo_snapshot saves the inventory of everything labelled app=openfief and the price list into $work.
demo_snapshot() {
  hcloud all list --selector app=openfief --output json >"$work/inventory.json"
  hcloud api /pricing >"$work/pricing.json"
}

# demo_has prints the number of resources of a kind (servers, firewalls) with this environment's label.
demo_has() {
  gojq -r --arg kind "$1" --arg env "$ENV" '[.[$kind][]? | select(.labels.env == $env)] | length' "$work/inventory.json"
}

# demo_plan prints what `up` would create and the cost arithmetic; it fails when creating would exceed the budget.
demo_plan() {
  local name="openfief-$ENV"
  demo_snapshot
  if [ "$(demo_has firewalls)" -eq 0 ]; then
    echo "would create: firewall $name (no charge)"
  fi
  if [ "$(demo_has servers)" -eq 0 ]; then
    echo "would create: server $name ($DEMO_SERVER_TYPE, $DEMO_IMAGE, $DEMO_LOCATION) with a primary IPv4"
    gojq -n -c --arg n "$name" --arg t "$DEMO_SERVER_TYPE" --arg l "$DEMO_LOCATION" \
      '{servers: [{name: $n, type: $t, location: $l}]}' >"$work/planned.json"
  else
    echo "server $name exists; nothing billable to create"
    echo '{"servers":[]}' >"$work/planned.json"
  fi
  echo "monthly cost of everything labelled app=openfief in the project, plus what would be created:"
  demo_cost_lines "$work/pricing.json" "$work/inventory.json" "$work/planned.json" >"$work/costs.tsv"
  if [ "$(gojq '.servers | length' "$work/planned.json")" -gt 0 ]; then
    demo_budget_check "$DEMO_BUDGET_EUR" <"$work/costs.tsv"
  else
    demo_budget_check "$DEMO_BUDGET_EUR" <"$work/costs.tsv" || echo "warning: over budget, but nothing new is billable" >&2
  fi
}

# demo_up creates the firewall and the server when missing, and brings the firewall rules up to date.
demo_up() {
  local name="openfief-$ENV" keys key args=() server_firewalls firewall_id ip
  hcloud ssh-key list --selector openfief-ssh=true --output json >"$work/keys.json"
  keys="$(demo_ssh_keys "$work/keys.json")"
  demo_firewall_rules "$DEMO_SSH_CIDRS" "$DEMO_HTTP_CIDRS" >"$work/rules.json"
  demo_render_cloud_init "$root/deploy/demo/cloud-init.yaml" "$DEMO_SERVER_TYPE" >"$work/user-data.yaml"
  if [ "$(gojq '[.[] | select(.port == "22")] | length' "$work/rules.json")" -eq 0 ]; then
    echo "note: DEMO_SSH_CIDRS is empty, so the firewall allows no SSH"
  fi
  # Under set -e, a failed plan (an API error or the budget) stops here, before anything is created.
  demo_plan
  if [ "$(demo_has firewalls)" -eq 0 ]; then
    hcloud firewall create --name "$name" --label app=openfief --label "env=$ENV" --rules-file "$work/rules.json"
  else
    hcloud firewall replace-rules --rules-file "$work/rules.json" "$name"
  fi
  if [ "$(demo_has servers)" -eq 0 ]; then
    for key in $keys; do args+=(--ssh-key "$key"); done
    hcloud server create --name "$name" --type "$DEMO_SERVER_TYPE" --image "$DEMO_IMAGE" \
      --location "$DEMO_LOCATION" --firewall "$name" "${args[@]}" \
      --user-data-from-file "$work/user-data.yaml" --label app=openfief --label "env=$ENV"
  else
    firewall_id="$(hcloud firewall describe "$name" --output json | gojq -r .id)"
    server_firewalls="$(gojq -r --arg n "$name" '.servers[] | select(.name == $n) | .public_net.firewalls[]?.id' \
      "$work/inventory.json")"
    if ! grep -qx "$firewall_id" <<<"$server_firewalls"; then
      hcloud firewall apply-to-resource --type server --server "$name" "$name"
    fi
    echo "server $name already exists; cloud-init changes need: task demo:recreate ENV=$ENV CONFIRM=$ENV"
  fi
  ip="$(hcloud server ip "$name")"
  if [ "$(demo_host)" = none ]; then
    echo "demo URL: https://${ip//./-}.sslip.io (TLS from Let's Encrypt $DEMO_ACME, once Flux and cert-manager are up)"
  else
    echo "demo URL: https://$(demo_host) (TLS from Let's Encrypt $DEMO_ACME, once Flux and cert-manager are up)"
    echo "DNS: $(demo_host) needs an A record to $ip, not proxied, so Let's Encrypt's HTTP-01 check reaches the machine"
  fi
  echo "next: docs/dev/demo.md#reach-the-demo (ssh root@$ip)"
}

# demo_down deletes every resource labelled app=openfief,env=<ENV>, servers first, after the name is confirmed.
demo_down() {
  if [ "${CONFIRM:-}" != "$ENV" ]; then
    demo_fail "down deletes the demo '$ENV' and its data; repeat with CONFIRM=$ENV"
    return 1
  fi
  local selector="app=openfief,env=$ENV" kind names
  hcloud all list --selector "$selector" --output json >"$work/env.json"
  for kind in server:servers firewall:firewalls primary-ip:primary_ips floating-ip:floating_ips volume:volumes; do
    names="$(gojq -r --arg k "${kind#*:}" '.[$k][]?.name' "$work/env.json")"
    if [ -n "$names" ]; then
      # shellcheck disable=SC2086 # one argument per resource name
      hcloud "${kind%%:*}" delete $names
    fi
  done
  echo "demo '$ENV' deleted"
}

# demo_status prints the environment's resources and the project's monthly cost.
demo_status() {
  demo_snapshot
  gojq -r --arg env "$ENV" '
    [.servers[]? | select(.labels.env == $env)
      | "server \(.name): \(.status), \(.server_type.name), \(.location.name), IPv4 \(.public_net.ipv4.ip // "none")"],
    [.firewalls[]? | select(.labels.env == $env)
      | "firewall \(.name): \([.rules[]? | select(.direction == "in")] | length) inbound rule(s)"]
    | .[]' "$work/inventory.json"
  [ "$(demo_has servers)" -gt 0 ] || echo "no server for demo '$ENV'"
  echo '{"servers":[]}' >"$work/planned.json"
  echo "monthly cost of everything labelled app=openfief in the project:"
  demo_cost_lines "$work/pricing.json" "$work/inventory.json" "$work/planned.json" >"$work/costs.tsv"
  demo_budget_check "$DEMO_BUDGET_EUR" <"$work/costs.tsv" || true
}

# demo_main dispatches a subcommand for the environment named by ENV.
demo_main() {
  local command="${1:-}"
  demo_check_env "${ENV:-}"
  demo_prepare
  case "$command" in
    plan) demo_plan ;;
    up) demo_up ;;
    status) demo_status ;;
    down) demo_down ;;
    recreate)
      demo_down
      demo_up
      ;;
    *)
      demo_fail "usage: ENV=<name> demo.sh plan|up|status|down|recreate"
      return 1
      ;;
  esac
}

# Run as a script; tests source this file for the pure helpers (demo_to_micro, demo_render_cloud_init, ...).
if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  work="$(mktemp -d)"
  trap 'rm -rf "$work"' EXIT
  demo_main "$@"
fi
