#!/usr/bin/env bash
# Applies deploy/grafana to the demo's Grafana Cloud stack (docs/dev/demo.md#grafana). Called through task:
#   task grafana:apply GRAFANA_URL=https://<stack>.grafana.net     with GRAFANA_CLOUD_TOKEN in the environment
# It upserts, in this order and by fixed uid, so a second run changes nothing:
#   the folder           deploy/grafana/folder.json                POST /api/folders, or PUT when it exists
#   the dashboards       deploy/grafana/dashboards/*.json          POST /api/dashboards/db with overwrite true
#   the alert rules      deploy/grafana/alerts/*.json              PUT /api/v1/provisioning/alert-rules/<uid>, POST on 404
# Alert rules are sent with X-Disable-Provenance, so they stay editable in Grafana's UI: the next apply replaces what
# was edited there, as it does a dashboard. It runs the offline check first. GRAFANA_URL has no default, and must be an
# https address on grafana.net: the token is a stack service account's, and goes nowhere else. The token reaches curl
# in a header file read from a process substitution (never an argument, never a file), and the script prints only
# what it changed, by name, and the HTTP status of each call: never a response body, a header or the token.
# Usage: apply.sh   (curl and gojq on PATH; GRAFANA_URL and GRAFANA_CLOUD_TOKEN in the environment)
set -euo pipefail
shopt -s inherit_errexit

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
dir="${GRAFANA_DIR:-$root/deploy/grafana}"

grafana_fail() {
  echo "grafana: $*" >&2
  return 1
}

# grafana_url prints the stack's address without a trailing slash, or fails: https, a host under grafana.net, no
# credentials, port or path.
grafana_url() {
  local url="${GRAFANA_URL:-}"
  if [ -z "$url" ]; then
    grafana_fail "GRAFANA_URL is not set: the stack's address, such as https://<stack>.grafana.net"
    return 1
  fi
  url="${url%/}"
  if [[ ! "$url" =~ ^https://[a-z0-9-]+([.][a-z0-9-]+)*[.]grafana[.]net$ ]]; then
    grafana_fail "GRAFANA_URL must be an https address on grafana.net, with no path, port or credentials: not used"
    return 1
  fi
  printf '%s' "$url"
}

# grafana_api <METHOD> <path> [<body file> [extra curl arguments]] calls the API and prints the HTTP status. The
# response body goes to $work/response and is never printed.
grafana_api() {
  local method="$1" path="$2" body="${3:-}" status
  local args=(-sS -o "$work/response" -w '%{http_code}' -X "$method" --max-time 60
    -H 'Content-Type: application/json' -H 'Accept: application/json')
  if [ -n "$body" ]; then
    args+=(--data-binary "@$body")
  fi
  if [ "${4:-}" = no-provenance ]; then
    args+=(-H 'X-Disable-Provenance: true')
  fi
  # The Authorization header lives only in the pipe of this process substitution.
  status="$(curl "${args[@]}" -H @<(printf 'Authorization: Bearer %s' "$GRAFANA_CLOUD_TOKEN") "$base$path")" || {
    echo "000"
    return 0
  }
  printf '%s' "$status"
}

grafana_ok() {
  [[ "$1" =~ ^2[0-9][0-9]$ ]]
}

# grafana_report <what> <status> prints one line and fails when the status is not a success.
grafana_report() {
  echo "$1: $2"
  grafana_ok "$2" || { grafana_fail "$1 failed with HTTP $2"; return 1; }
}

grafana_folder() {
  local uid title status
  uid="$(gojq -r .uid "$dir/folder.json")"
  title="$(gojq -r .title "$dir/folder.json")"
  status="$(grafana_api GET "/api/folders/$uid")"
  if [ "$status" = 404 ]; then
    status="$(grafana_api POST /api/folders "$dir/folder.json")"
    grafana_report "folder $uid (create)" "$status"
  else
    gojq -c --arg title "$title" '{title: $title, overwrite: true}' "$dir/folder.json" >"$work/folder-body.json"
    status="$(grafana_api PUT "/api/folders/$uid" "$work/folder-body.json")"
    grafana_report "folder $uid (update)" "$status"
  fi
}

grafana_dashboards() {
  local file uid status
  for file in "$dir"/dashboards/*.json; do
    uid="$(gojq -r .uid "$file")"
    gojq -c --arg folder "$(gojq -r .uid "$dir/folder.json")" \
      '{dashboard: (. + {id: null}), folderUid: $folder, overwrite: true, message: "task grafana:apply"}' "$file" >"$work/dashboard-body.json"
    status="$(grafana_api POST /api/dashboards/db "$work/dashboard-body.json")"
    grafana_report "dashboard $uid" "$status"
  done
}

grafana_rules() {
  local file uid status
  for file in "$dir"/alerts/*.json; do
    uid="$(gojq -r .uid "$file")"
    status="$(grafana_api PUT "/api/v1/provisioning/alert-rules/$uid" "$file" no-provenance)"
    if [ "$status" = 404 ]; then
      status="$(grafana_api POST /api/v1/provisioning/alert-rules "$file" no-provenance)"
      grafana_report "alert rule $uid (create)" "$status"
    else
      grafana_report "alert rule $uid (update)" "$status"
    fi
  done
}

grafana_main() {
  local tool
  for tool in curl gojq; do
    command -v "$tool" >/dev/null 2>&1 || grafana_fail "$tool not found" || return 1
  done
  base="$(grafana_url)"
  if [ -z "${GRAFANA_CLOUD_TOKEN:-}" ]; then
    grafana_fail "GRAFANA_CLOUD_TOKEN is not set: a service account token of the stack with the Admin role (read -rs GRAFANA_CLOUD_TOKEN && export GRAFANA_CLOUD_TOKEN)"
    return 1
  fi
  bash "$here/check.sh" "$dir"
  work="$(mktemp -d)"
  trap 'rm -rf "$work"' EXIT
  grafana_folder
  grafana_dashboards
  grafana_rules
  echo "grafana: applied ${dir#"$root"/} to $base"
}

grafana_main "$@"
