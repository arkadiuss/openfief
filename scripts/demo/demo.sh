#!/usr/bin/env bash
# Creates, inspects and destroys a demo machine on Hetzner Cloud (docs/dev/demo.md). Called through task:
#   task demo:plan ENV=demo        what `up` would create and the monthly cost; changes nothing
#   task demo:up ENV=demo          create what is missing (idempotent), within DEMO_BUDGET_EUR
#   task demo:status ENV=demo      the environment's resources and cost
#   task demo:down ENV=demo CONFIRM=demo      delete every resource labelled app=openfief,env=demo
#   task demo:recreate ENV=demo CONFIRM=demo  down, then up
#   task dns:docs:plan / task dns:docs        the docs site's CNAME docs.<DEMO_DOMAIN>
# Stateless: every resource carries the labels app=openfief and env=<ENV> and is found by them, and every DNS
# record the comment "managed-by=openfief-demo env=<ENV>".
# The Hetzner token comes from HCLOUD_TOKEN (or HETZNER_API_TOKEN), the Cloudflare token from CLOUDFLARE_API_TOKEN;
# neither is printed, written to disk or passed as a command argument. Without CLOUDFLARE_API_TOKEN, the demo
# has no DNS record and is reached at <ipv4-with-dashes>.sslip.io.
set -euo pipefail
# A failing command inside $(...) fails the substitution too, as it would outside.
shopt -s inherit_errexit

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
or label an existing key: hcloud ssh-key add-label <key> openfief-ssh=true
or create the demo without SSH: DEMO_SSH=off, with DEMO_SSH_CIDRS empty"
    return 1
  fi
  echo "$names"
}

# demo_ssh_mode prints how a new server gets its SSH key: "keys" installs the project's keys labelled
# openfief-ssh=true, "off" a throwaway key whose private half nobody has. The arguments are the DEMO_SSH setting
# (auto, keys or off), the labelled keys (an `hcloud ssh-key list -o json` file) and DEMO_SSH_CIDRS. auto picks
# keys when a labelled key exists, off when none does and DEMO_SSH_CIDRS is empty, and refuses otherwise.
demo_ssh_mode() {
  local setting="$1" keys="$2" cidrs count
  cidrs="$(demo_cidrs DEMO_SSH_CIDRS "$3")" || return 1
  count="$(gojq 'length' "$keys")"
  case "$setting" in
    auto)
      if [ "$count" -gt 0 ]; then
        echo keys
      elif [ "$cidrs" = '[]' ]; then
        echo off
      else
        demo_ssh_keys "$keys" >/dev/null
        return 1
      fi
      ;;
    keys)
      demo_ssh_keys "$keys" >/dev/null || return 1
      echo keys
      ;;
    off)
      if [ "$cidrs" != '[]' ]; then
        demo_fail "DEMO_SSH=off allows no SSH, so DEMO_SSH_CIDRS must be empty (it is '$3')"
        return 1
      fi
      echo off
      ;;
    *)
      demo_fail "unknown DEMO_SSH '$setting', want auto, keys or off"
      return 1
      ;;
  esac
}

# demo_throwaway_pubkey prints a new Ed25519 public key in OpenSSH form, with the comment in the second argument.
# The private key exists only inside the openssl pipe and is never written anywhere, so no one can log in with
# the key; installing it only stops Hetzner from setting a root password and emailing it. The first argument is a
# scratch file for the public key in DER form.
demo_throwaway_pubkey() {
  local der="$1" comment="$2" blob
  if ! command -v openssl >/dev/null 2>&1; then
    demo_fail "openssl not found; DEMO_SSH=off needs it to make a throwaway key"
    return 1
  fi
  openssl genpkey -algorithm ed25519 2>/dev/null | openssl pkey -pubout -outform DER >"$der"
  # An Ed25519 SubjectPublicKeyInfo is 44 bytes: a 12-byte header, then the 32-byte key.
  if [ "$(($(wc -c <"$der")))" -ne 44 ]; then
    demo_fail "openssl made no Ed25519 public key"
    return 1
  fi
  # OpenSSH's wire form: the string "ssh-ed25519" (length 11), then the key as a string (length 32).
  blob="$({ printf '\000\000\000\013ssh-ed25519\000\000\000\040' && tail -c 32 "$der"; } | base64 | tr -d '\n')"
  printf 'ssh-ed25519 %s %s\n' "$blob" "$comment"
}

# demo_host_form matches a host name the scripts accept: lowercase DNS labels, the last one alphabetic.
demo_host_form='^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$'

# demo_host prints the demo's host name, <env>.<domain> (the first and second arguments), such as
# demo.openfief.dev, or "none" when the domain is none: the machine then names itself <ipv4-with-dashes>.sslip.io
# at boot. It fails on a domain that is not a DNS name.
demo_host() {
  local env="$1" domain="$2"
  if [ "$domain" = none ]; then
    echo none
    return 0
  fi
  if [[ ! "$env.$domain" =~ $demo_host_form ]]; then
    demo_fail "DEMO_DOMAIN '$domain' is not a lowercase DNS name such as example.com (or none for sslip.io)"
    return 1
  fi
  echo "$env.$domain"
}

# demo_dns_comment prints the comment that marks a DNS record as the demo's: the DNS counterpart of the labels.
demo_dns_comment() {
  echo "managed-by=openfief-demo env=$1"
}

# demo_dns_desired prints, as a JSON array, the DNS records a demo needs: an A record to its IPv4, an AAAA record
# to its IPv6 when one is given, and a CAA record that lets only Let's Encrypt issue certificates for the host.
# Every record carries the demo's comment, and none is proxied: Let's Encrypt's HTTP-01 check and TLS end on the
# machine itself.
demo_dns_desired() {
  gojq -n -c --arg host "$1" --arg ipv4 "$2" --arg ipv6 "$3" --arg comment "$4" '
    [{type: "A", name: $host, content: $ipv4, ttl: 60, proxied: false, comment: $comment},
     (if $ipv6 != "" then {type: "AAAA", name: $host, content: $ipv6, ttl: 60, proxied: false, comment: $comment}
      else empty end),
     {type: "CAA", name: $host, ttl: 60, data: {flags: 0, tag: "issue", value: "letsencrypt.org"}, comment: $comment}]'
}

# demo_dns_changes compares the desired records (a file from demo_dns_desired) with the zone's existing ones (a
# file holding a JSON array of Cloudflare records) and prints one JSON action per line: create, update or delete.
# Only records whose comment equals the third argument are ever updated or deleted. It fails, printing nothing,
# when a record without that comment would clash at the host in the fourth argument (A, AAAA, CNAME or CAA;
# any type when a CNAME is wanted, since a CNAME must stand alone).
demo_dns_changes() {
  local actions clash
  actions="$(gojq -c -n --slurpfile desired "$1" --slurpfile existing "$2" --arg comment "$3" --arg host "$4" '
    def key: "\(.type) \(.name)";
    def same($d): if .type == "CAA"
      then (.data.flags == $d.data.flags and .data.tag == $d.data.tag and .data.value == $d.data.value)
      else (.content == $d.content and (.proxied // false) == false) end;
    $desired[0] as $want
    | [$existing[0][] | select(.comment == $comment)] as $ours
    | ([$existing[0][] | select(.comment != $comment and .name == $host
          and ((.type | IN("A", "AAAA", "CNAME", "CAA")) or ($want | any(.type == "CNAME"))))
        | {op: "clash", record: .}]
       + [$want[] as $d
          | ([$ours[] | select(key == ($d | key))] | .[0]) as $e
          | if $e == null then {op: "create", record: $d}
            elif ($e | same($d)) then empty
            else {op: "update", id: $e.id, from: $e, record: $d} end]
       + [$ours | group_by(key)[] as $g
          | (if ($want | any(key == ($g[0] | key))) then $g[1:] else $g end)[]
          | {op: "delete", id: .id, record: .}])[]')" || return 1
  clash="$(gojq -r 'select(.op == "clash") | .record | "\(.type) \(.name)"' <<<"$actions")"
  if [ -n "$clash" ]; then
    demo_fail "refusing: the zone has $(echo "$clash" | paste -sd, -) without the comment \"$3\"; remove it, or give it that comment so these scripts manage it"
    return 1
  fi
  [ -z "$actions" ] || printf '%s\n' "$actions"
}

# demo_dns_show prints DNS records or actions (JSON, one per line on stdin) as text, one per line.
demo_dns_show() {
  gojq -r '
    def value: if .type == "CAA" then "\(.data.flags) \(.data.tag) \"\(.data.value)\"" else .content end;
    if .op == null then "\(.type) \(.name) \(value)"
    elif .op == "update" then "update \(.record.type) \(.record.name) \(.from | value) -> \(.record | value)"
    else "\(.op) \(.record.type) \(.record.name) \(.record | value)" end'
}

# demo_ipv6_host prints the address Hetzner gives a server inside its IPv6 network, such as 2001:db8::1 for
# 2001:db8::/64, or nothing when the server has no IPv6.
demo_ipv6_host() {
  local net="${1%/*}"
  case "$net" in
    "" | null) ;;
    *::) echo "${net}1" ;;
    *) demo_fail "unexpected IPv6 network '$1'" ;;
  esac
}

# demo_render_cloud_init fills the @@NAME@@ placeholders of a cloud-init template from the pinned versions in the
# environment, after checking each value's form, and fails if a placeholder is left. The third argument is the
# demo's host name (demo_host), or "none" (the default) for the machine to use <ipv4-with-dashes>.sslip.io.
demo_render_cloud_init() {
  local template="$1" server_type="$2" host="${3:-none}" asset k3s_sha acme rendered name value
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
    [ACME_SERVER]="$acme" [DEMO_HOST]="$host"
    [K3S_VERSION]="${K3S_VERSION:-}" [K3S_ASSET]="$asset" [K3S_SHA256]="$k3s_sha"
    [K3S_INSTALL_SHA256]="${K3S_INSTALL_SHA256:-}" [FLUX_VERSION]="${FLUX_VERSION:-}"
    [FLUX_INSTALL_SHA256]="${FLUX_INSTALL_SHA256:-}"
    [GIT_URL]="${DEMO_GIT_URL:-}" [GIT_BRANCH]="${DEMO_GIT_BRANCH:-}"
  )
  local -A forms=(
    [ACME_SERVER]='^https://acme(-staging)?-v02\.api\.letsencrypt\.org/directory$'
    [DEMO_HOST]="^none\$|${demo_host_form}"
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
: "${DEMO_HTTP_CIDRS:=0.0.0.0/0,::/0}" "${DEMO_ACME:=prod}" "${DEMO_SSH:=auto}"
# DNS, used when CLOUDFLARE_API_TOKEN is set. k3s on the demo is IPv4-only, so an AAAA record would send IPv6
# clients, Let's Encrypt among them, to a port nothing listens on: DEMO_DNS_AAAA stays off until k3s serves IPv6.
: "${DEMO_DOMAIN:=openfief.dev}" "${DEMO_DNS_AAAA:=off}"
export DEMO_GIT_URL DEMO_GIT_BRANCH DEMO_ACME
cf_api=https://api.cloudflare.com/client/v4

# demo_prepare checks the tokens, settings and tools, isolates hcloud from any local CLI context, and sets the
# demo's host name (host: <ENV>.<DEMO_DOMAIN>, or none for sslip.io), whether the scripts manage its DNS records
# (dns: on with a domain and CLOUDFLARE_API_TOKEN) and the comment those records carry.
demo_prepare() {
  # HETZNER_API_TOKEN is accepted as another name for the same token, as some sessions call it.
  if [ -z "${HCLOUD_TOKEN:-}" ] && [ -n "${HETZNER_API_TOKEN:-}" ]; then
    HCLOUD_TOKEN="$HETZNER_API_TOKEN"
  fi
  if [ -z "${HCLOUD_TOKEN:-}" ]; then
    demo_fail "HCLOUD_TOKEN is not set (nor HETZNER_API_TOKEN). Create a Read & Write API token in the Hetzner Cloud
console (project > Security > API tokens) and export it in this shell only, e.g. read -rs HCLOUD_TOKEN && export HCLOUD_TOKEN"
    return 1
  fi
  export HCLOUD_TOKEN
  local tool tools=(hcloud gojq)
  if [ -n "${CLOUDFLARE_API_TOKEN:-}" ] && [ "$DEMO_DOMAIN" != none ]; then tools+=(curl); fi
  for tool in "${tools[@]}"; do
    command -v "$tool" >/dev/null 2>&1 || demo_fail "$tool not found; run task setup" || return 1
  done
  case "$DEMO_DNS_AAAA" in
    on | off) ;;
    *) demo_fail "unknown DEMO_DNS_AAAA '$DEMO_DNS_AAAA', want on or off" || return 1 ;;
  esac
  # A missing config file: hcloud uses HCLOUD_TOKEN, never a stored context of another project.
  export HCLOUD_CONFIG="$work/no-cli.toml"
  unset HCLOUD_CONTEXT
  comment="$(demo_dns_comment "$ENV")"
  host="$(demo_host "$ENV" "$DEMO_DOMAIN")"
  dns=off
  if [ "$host" != none ] && [ -n "${CLOUDFLARE_API_TOKEN:-}" ]; then
    dns=on
  elif [ "$host" != none ]; then
    echo "note: CLOUDFLARE_API_TOKEN is not set, so the scripts manage no DNS record; up prints the one $host needs"
  fi
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

# demo_cf calls the Cloudflare API with a method, a path under /client/v4 and an optional JSON body file, and
# prints the response. The token reaches curl on stdin as a header, never as an argument, so no process list or
# log shows it. It fails when the API reports an error, printing only the API's error codes and messages.
demo_cf() {
  local method="$1" path="$2" body="${3:-}" response errors
  local args=(-sS --max-time 60 -X "$method" -H @- -H 'Content-Type: application/json')
  if [ -n "$body" ]; then args+=(--data-binary "@$body"); fi
  if ! response="$(printf 'Authorization: Bearer %s\n' "$CLOUDFLARE_API_TOKEN" | curl "${args[@]}" "$cf_api$path")"; then
    demo_fail "Cloudflare API: $method request failed"
    return 1
  fi
  if ! printf '%s' "$response" | gojq -e '.success == true' >/dev/null 2>&1; then
    errors="$(printf '%s' "$response" | gojq -r '[.errors[]? | "\(.code) \(.message)"] | join("; ")' 2>/dev/null)" ||
      errors='unreadable response'
    demo_fail "Cloudflare API: $method failed: $errors"
    return 1
  fi
  printf '%s\n' "$response"
}

# demo_dns_zone sets zone to the id of the Cloudflare zone DEMO_DOMAIN, once.
demo_dns_zone() {
  [ -z "${zone:-}" ] || return 0
  demo_cf GET "/zones?name=$DEMO_DOMAIN" >"$work/zones.json"
  zone="$(gojq -r --arg d "$DEMO_DOMAIN" '[.result[]? | select(.name == $d) | .id] | .[0] // ""' "$work/zones.json")"
  if [[ ! "$zone" =~ ^[0-9a-f]{32}$ ]]; then
    demo_fail "CLOUDFLARE_API_TOKEN can't see the zone $DEMO_DOMAIN; the token needs Zone > Zone > Read and
Zone > DNS > Edit on that zone (docs/dev/demo.md#dns)"
    return 1
  fi
}

# demo_dns_records saves into $work/dns.json, as one JSON array, the zone's records that carry this demo's
# comment or sit at its host name. The comment is checked again here, so a filter the API ignores can't widen it.
demo_dns_records() {
  local query
  demo_dns_zone
  query="$(gojq -rn --arg c "$comment" '$c | @uri')"
  demo_cf GET "/zones/$zone/dns_records?per_page=100&comment.exact=$query" >"$work/dns-ours.json"
  demo_cf GET "/zones/$zone/dns_records?per_page=100&name.exact=$host" >"$work/dns-host.json"
  if ! gojq -e -s 'all(.[]; (.result_info.total_count // 0) <= (.result | length))' \
    "$work/dns-ours.json" "$work/dns-host.json" >/dev/null; then
    demo_fail "more than 100 DNS records match the demo's host or comment; clean up the zone first"
    return 1
  fi
  gojq -s -c --arg c "$comment" --arg h "$host" '[.[].result[] | select(.comment == $c or .name == $h)] | unique_by(.id)' \
    "$work/dns-ours.json" "$work/dns-host.json" >"$work/dns.json"
}

# demo_dns_plan_for saves into $work/dns-changes.jsonl what DNS needs to change for the server addresses given.
demo_dns_plan_for() {
  demo_dns_records
  demo_dns_desired "$host" "$1" "$2" "$comment" >"$work/dns-desired.json"
  demo_dns_changes "$work/dns-desired.json" "$work/dns.json" "$comment" "$host" >"$work/dns-changes.jsonl"
}

# demo_dns_plan prints what `up` would change in DNS, using the existing server's addresses, if any.
demo_dns_plan() {
  [ "$dns" = on ] || return 0
  local ipv4 ipv6=""
  gojq -c --arg env "$ENV" '[.servers[]? | select(.labels.env == $env)] | .[0] // {}' "$work/inventory.json" \
    >"$work/planned-server.json"
  ipv4="$(gojq -r '.public_net.ipv4.ip // "<IPv4 of the new server>"' "$work/planned-server.json")"
  if [ "$DEMO_DNS_AAAA" = on ]; then
    ipv6="$(demo_ipv6_host "$(gojq -r '.public_net.ipv6.ip // ""' "$work/planned-server.json")")"
    [ -n "$ipv6" ] || ipv6="<IPv6 of the new server>"
  fi
  demo_dns_plan_for "$ipv4" "$ipv6"
  if [ -s "$work/dns-changes.jsonl" ]; then
    demo_dns_show <"$work/dns-changes.jsonl" | sed 's/^/would DNS: /'
  else
    echo "DNS records of $host are up to date"
  fi
}

# demo_dns_apply creates, updates and deletes this demo's DNS records to match the server described in the file
# given (hcloud server describe -o json).
demo_dns_apply() {
  [ "$dns" = on ] || return 0
  local ipv4 ipv6=""
  ipv4="$(gojq -r '.public_net.ipv4.ip // ""' "$1")"
  if [ -z "$ipv4" ]; then
    demo_fail "the server has no public IPv4 for the A record"
    return 1
  fi
  if [ "$DEMO_DNS_AAAA" = on ]; then
    ipv6="$(demo_ipv6_host "$(gojq -r '.public_net.ipv6.ip // ""' "$1")")"
  fi
  demo_dns_plan_for "$ipv4" "$ipv6"
  demo_dns_apply_changes
}

# demo_dns_apply_changes applies the actions in $work/dns-changes.jsonl (from demo_dns_changes) in the zone, and
# refuses to update or delete any record whose comment isn't the one in the global comment.
demo_dns_apply_changes() {
  local action op id
  while IFS= read -r action; do
    op="$(gojq -r .op <<<"$action")"
    id="$(gojq -r '.id // ""' <<<"$action")"
    gojq -c .record <<<"$action" >"$work/dns-record.json"
    if [ "$op" != create ] && [[ ! "$id" =~ ^[0-9a-f]{32}$ ]]; then
      demo_fail "unexpected DNS record id from Cloudflare"
      return 1
    fi
    # A second guard: nothing without our comment is ever changed or deleted.
    if [ "$op" != create ] && [ "$(gojq -r '(.from // .record).comment // ""' <<<"$action")" != "$comment" ]; then
      demo_fail "refusing to $op a DNS record without the comment \"$comment\""
      return 1
    fi
    case "$op" in
      create) demo_cf POST "/zones/$zone/dns_records" "$work/dns-record.json" >/dev/null ;;
      update) demo_cf PATCH "/zones/$zone/dns_records/$id" "$work/dns-record.json" >/dev/null ;;
      delete) demo_cf DELETE "/zones/$zone/dns_records/$id" >/dev/null ;;
    esac
    echo "DNS: $(demo_dns_show <<<"$action")"
  done <"$work/dns-changes.jsonl"
}

# demo_dns_down deletes every DNS record that carries this demo's comment, and nothing else.
demo_dns_down() {
  [ "$host" != none ] || return 0
  if [ "$dns" != on ]; then
    echo "note: without CLOUDFLARE_API_TOKEN, any DNS records of demo '$ENV' stay; run down again with it to delete them"
    return 0
  fi
  local id
  demo_dns_records
  gojq -c --arg c "$comment" '.[] | select(.comment == $c)' "$work/dns.json" >"$work/dns-ours.jsonl"
  while IFS= read -r record; do
    id="$(gojq -r .id <<<"$record")"
    if [ "$(gojq -r '.comment // ""' <<<"$record")" != "$comment" ]; then
      demo_fail "refusing to delete a DNS record without the comment \"$comment\""
      return 1
    fi
    if [[ ! "$id" =~ ^[0-9a-f]{32}$ ]]; then
      demo_fail "unexpected DNS record id from Cloudflare"
      return 1
    fi
    demo_cf DELETE "/zones/$zone/dns_records/$id" >/dev/null
    echo "DNS: delete $(demo_dns_show <<<"$record")"
  done <"$work/dns-ours.jsonl"
}

# demo_dns_status prints this demo's DNS records.
demo_dns_status() {
  [ "$dns" = on ] || return 0
  demo_dns_records
  gojq -c --arg c "$comment" '.[] | select(.comment == $c)' "$work/dns.json" >"$work/dns-ours.jsonl"
  if [ -s "$work/dns-ours.jsonl" ]; then
    demo_dns_show <"$work/dns-ours.jsonl" | sed 's/^/DNS record /'
  else
    echo "no DNS record for demo '$ENV' (comment \"$comment\")"
  fi
}

# demo_url prints the demo's URL for the server's IPv4 in the first argument.
demo_url() {
  if [ "$host" = none ]; then echo "https://${1//./-}.sslip.io"; else echo "https://$host"; fi
}

# demo_ssh_plan prints how a new server would get SSH access, and fails when up would refuse.
demo_ssh_plan() {
  local mode
  hcloud ssh-key list --selector openfief-ssh=true --output json >"$work/keys.json"
  mode="$(demo_ssh_mode "$DEMO_SSH" "$work/keys.json" "$DEMO_SSH_CIDRS")"
  if [ "$mode" = keys ]; then
    echo "SSH keys for a new server: $(demo_ssh_keys "$work/keys.json" | paste -sd, -)"
  else
    echo "SSH off: a new server gets a throwaway key whose private half is never kept (DEMO_SSH=off)"
  fi
}

# demo_throwaway_key prints the name of this demo's throwaway SSH key, uploading a new one when none exists. Only
# the public half ever leaves openssl; reusing a key is safe because nobody holds its private half.
demo_throwaway_key() {
  local name="openfief-$ENV-throwaway" existing
  hcloud ssh-key list --selector "app=openfief,env=$ENV,openfief-ssh=throwaway" --output json >"$work/throwaway.json"
  existing="$(gojq -r '.[0].name // ""' "$work/throwaway.json")"
  if [ -n "$existing" ]; then
    echo "$existing"
    return 0
  fi
  demo_throwaway_pubkey "$work/throwaway.der" "$name" >"$work/throwaway.pub"
  hcloud ssh-key create --name "$name" --public-key-from-file "$work/throwaway.pub" \
    --label app=openfief --label "env=$ENV" --label openfief-ssh=throwaway >&2
  echo "$name"
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
    demo_ssh_plan
    gojq -n -c --arg n "$name" --arg t "$DEMO_SERVER_TYPE" --arg l "$DEMO_LOCATION" \
      '{servers: [{name: $n, type: $t, location: $l}]}' >"$work/planned.json"
  else
    echo "server $name exists; nothing billable to create"
    echo '{"servers":[]}' >"$work/planned.json"
  fi
  demo_dns_plan
  echo "monthly cost of everything labelled app=openfief in the project, plus what would be created:"
  demo_cost_lines "$work/pricing.json" "$work/inventory.json" "$work/planned.json" >"$work/costs.tsv"
  if [ "$(gojq '.servers | length' "$work/planned.json")" -gt 0 ]; then
    demo_budget_check "$DEMO_BUDGET_EUR" <"$work/costs.tsv"
  else
    demo_budget_check "$DEMO_BUDGET_EUR" <"$work/costs.tsv" || echo "warning: over budget, but nothing new is billable" >&2
  fi
}

# demo_up creates the firewall and the server when missing, brings the firewall rules up to date, and points the
# demo's DNS records at the server.
demo_up() {
  local name="openfief-$ENV" ssh key args=() server_firewalls firewall_id ip
  hcloud ssh-key list --selector openfief-ssh=true --output json >"$work/keys.json"
  ssh="$(demo_ssh_mode "$DEMO_SSH" "$work/keys.json" "$DEMO_SSH_CIDRS")"
  demo_firewall_rules "$DEMO_SSH_CIDRS" "$DEMO_HTTP_CIDRS" >"$work/rules.json"
  demo_render_cloud_init "$root/deploy/demo/cloud-init.yaml" "$DEMO_SERVER_TYPE" "$host" >"$work/user-data.yaml"
  if [ "$(gojq '[.[] | select(.port == "22")] | length' "$work/rules.json")" -eq 0 ]; then
    echo "note: DEMO_SSH_CIDRS is empty, so the firewall allows no SSH"
  fi
  # Under set -e, a failed plan (an API error, a DNS clash or the budget) stops here, before anything is created.
  demo_plan
  if [ "$(demo_has firewalls)" -eq 0 ]; then
    hcloud firewall create --name "$name" --label app=openfief --label "env=$ENV" --rules-file "$work/rules.json"
  else
    hcloud firewall replace-rules --rules-file "$work/rules.json" "$name"
  fi
  if [ "$(demo_has servers)" -eq 0 ]; then
    if [ "$ssh" = keys ]; then
      for key in $(demo_ssh_keys "$work/keys.json"); do args+=(--ssh-key "$key"); done
    else
      # Without any key, Hetzner would set a root password and email it.
      key="$(demo_throwaway_key)"
      args+=(--ssh-key "$key")
    fi
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
  hcloud server describe "$name" --output json >"$work/server.json"
  demo_dns_apply "$work/server.json"
  ip="$(gojq -r '.public_net.ipv4.ip' "$work/server.json")"
  echo "demo URL: $(demo_url "$ip") (TLS from Let's Encrypt $DEMO_ACME, once Flux and cert-manager are up)"
  if [ "$host" != none ] && [ "$dns" != on ]; then
    echo "DNS: $host needs an A record to $ip, not proxied, so Let's Encrypt's HTTP-01 check reaches the machine"
  fi
  if [ "$ssh" = keys ]; then
    echo "next: docs/dev/demo.md#reach-the-demo (ssh root@$ip)"
  else
    echo "SSH is off: nobody holds a key to the machine; it changes only through git and Flux (docs/dev/demo.md#ssh)"
  fi
}

# demo_down deletes every resource labelled app=openfief,env=<ENV>, servers first, then the demo's throwaway SSH
# key and, unless the first argument is keep-dns (recreate), its DNS records, after the name is confirmed.
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
  # Only throwaway keys carry the env label; the owner's keys labelled openfief-ssh=true are never deleted.
  hcloud ssh-key list --selector "$selector,openfief-ssh=throwaway" --output json >"$work/throwaway.json"
  names="$(gojq -r '.[].name' "$work/throwaway.json")"
  if [ -n "$names" ]; then
    # shellcheck disable=SC2086 # one argument per key name
    hcloud ssh-key delete $names
  fi
  if [ "${1:-}" = keep-dns ]; then
    echo "DNS records of demo '$ENV' kept for the new server"
  else
    demo_dns_down
  fi
  echo "demo '$ENV' deleted"
}

# demo_status prints the environment's resources, its DNS records and the project's monthly cost.
demo_status() {
  demo_snapshot
  gojq -r --arg env "$ENV" '
    [.servers[]? | select(.labels.env == $env)
      | "server \(.name): \(.status), \(.server_type.name), \(.location.name), IPv4 \(.public_net.ipv4.ip // "none")"],
    [.firewalls[]? | select(.labels.env == $env)
      | "firewall \(.name): \([.rules[]? | select(.direction == "in")] | length) inbound rule(s)"]
    | .[]' "$work/inventory.json"
  [ "$(demo_has servers)" -gt 0 ] || echo "no server for demo '$ENV'"
  demo_dns_status
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
      # The DNS records stay and are pointed at the new server, so the host name never disappears.
      demo_down keep-dns
      demo_up
      ;;
    *)
      demo_fail "usage: ENV=<name> demo.sh plan|up|status|down|recreate"
      return 1
      ;;
  esac
}

# The docs site's DNS record (task dns:docs, docs/dev/demo.md#dns): docs.<DEMO_DOMAIN> as a DNS-only CNAME to
# GitHub Pages with the comment "managed-by=openfief-docs". It reuses the Cloudflare helpers above and, like them,
# updates or deletes only a record with its own comment; the demo's records carry another, so neither touches the
# other's.
# The GitHub Pages host the docs site is published on.
: "${DOCS_PAGES_HOST:=arkadiuss.github.io}"

# docs_dns_desired prints the docs site's records as a JSON array: one DNS-only CNAME from the host in the first
# argument to the Pages host in the second, with the comment in the third. GitHub Pages issues its own certificate,
# which needs the name to resolve to GitHub directly, so the record is not proxied.
docs_dns_desired() {
  gojq -n -c --arg host "$1" --arg target "$2" --arg comment "$3" '
    [{type: "CNAME", name: $host, content: $target, ttl: 60, proxied: false, comment: $comment}]'
}


# docs_dns_main runs plan or apply for the docs site's record.
docs_dns_main() {
  local command="${1:-}" tool
  case "$command" in
    plan | apply) ;;
    *)
      demo_fail "usage: demo.sh dns-docs plan|apply"
      return 1
      ;;
  esac
  if [ -z "${CLOUDFLARE_API_TOKEN:-}" ]; then
    demo_fail "CLOUDFLARE_API_TOKEN is not set; it needs Zone > Zone > Read and Zone > DNS > Edit on $DEMO_DOMAIN"
    return 1
  fi
  if [ "$DEMO_DOMAIN" = none ]; then
    demo_fail "DEMO_DOMAIN=none: there is no zone for the docs site's record"
    return 1
  fi
  for tool in curl gojq; do
    command -v "$tool" >/dev/null 2>&1 || demo_fail "$tool not found; run task setup" || return 1
  done
  host="docs.$DEMO_DOMAIN"
  comment="managed-by=openfief-docs"
  if [[ ! "$host" =~ $demo_host_form ]] || [[ ! "$DOCS_PAGES_HOST" =~ $demo_host_form ]]; then
    demo_fail "'$host' or DOCS_PAGES_HOST '$DOCS_PAGES_HOST' is not a lowercase DNS name"
    return 1
  fi
  demo_dns_records
  docs_dns_desired "$host" "$DOCS_PAGES_HOST" "$comment" >"$work/dns-desired.json"
  demo_dns_changes "$work/dns-desired.json" "$work/dns.json" "$comment" "$host" >"$work/dns-changes.jsonl"
  if [ ! -s "$work/dns-changes.jsonl" ]; then
    echo "DNS record of $host is up to date (CNAME to $DOCS_PAGES_HOST)"
  elif [ "$command" = plan ]; then
    demo_dns_show <"$work/dns-changes.jsonl" | sed 's/^/would DNS: /'
  else
    demo_dns_apply_changes
  fi
}


# Run as a script, "demo.sh <command>" or "demo.sh dns-docs plan|apply"; tests source this file for the pure
# helpers (demo_to_micro, demo_render_cloud_init, ...).
if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  work="$(mktemp -d)"
  trap 'rm -rf "$work"' EXIT
  if [ "${1:-}" = dns-docs ]; then
    docs_dns_main "${2:-}"
  else
    demo_main "$@"
  fi
fi
