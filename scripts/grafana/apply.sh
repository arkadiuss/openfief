#!/usr/bin/env bash
# Applies deploy/grafana to the demo's Grafana Cloud stack (docs/dev/demo.md#grafana). Called through task:
#   task grafana:apply GRAFANA_URL=https://<stack>.grafana.net     with GRAFANA_CLOUD_TOKEN in the environment
# It upserts, in this order and by fixed uid, so a second run changes nothing:
#   the folder           deploy/grafana/folder.json                POST /api/folders, or PUT when it exists
#   the dashboards       deploy/grafana/dashboards/*.json          POST /api/dashboards/db with overwrite true
#   the alert rules      deploy/grafana/alerts/*.json              PUT /api/v1/provisioning/alert-rules/<uid>, POST on 404
#   the notifications    deploy/grafana/notifications.json         a Grafana IRM integration (OnCall API) whose default
#                        route is a Slack channel, a contact point of type oncall at the integration's link, and the
#                        folder's route in the notification policy tree (docs/dev/demo.md#grafana)
# Alert rules are sent with X-Disable-Provenance, so they stay editable in Grafana's UI: the next apply replaces what
# was edited there, as it does a dashboard. It runs the offline check first. GRAFANA_URL has no default, and must be an
# https address on grafana.net: the token is a stack service account's, and goes nowhere else. The token reaches curl
# in a header file read from a process substitution (never an argument, never a file), and the script prints only
# what it changed, by name, and the HTTP status of each call: never a response body, a header or the token. The
# integration's link is a credential too: it is read from the OnCall API's answer into a shell variable (memory only,
# never a file or an argument) and goes straight into the contact point's request, on stdin.
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

# grafana_mem <METHOD> <url> <bearer|raw> [<body>] calls the API with the answer and the body in memory only (the
# OnCall API's answers hold the integration's link) and sets mem_status and mem_body. "bearer" is the stack's own API;
# "raw" is the OnCall API's form: the token alone as Authorization, and the stack's address in X-Grafana-URL.
mem_status='' mem_body=''
grafana_mem() {
  local method="$1" url="$2" mode="$3" body="${4:-}" out
  local args=(-sS -w '\n%{http_code}' -X "$method" --max-time 60 -H 'Content-Type: application/json' -H 'Accept: application/json')
  if [ "$mode" = raw ]; then
    args+=(-H "X-Grafana-URL: $base")
    if [ -n "$body" ]; then
      out="$(printf '%s' "$body" | curl "${args[@]}" --data-binary @- -H @<(printf 'Authorization: %s' "$GRAFANA_CLOUD_TOKEN") "$url")" || out=$'\n000'
    else
      out="$(curl "${args[@]}" -H @<(printf 'Authorization: %s' "$GRAFANA_CLOUD_TOKEN") "$url" </dev/null)" || out=$'\n000'
    fi
  else
    if [ -n "$body" ]; then
      out="$(printf '%s' "$body" | curl "${args[@]}" --data-binary @- -H 'X-Disable-Provenance: true' -H @<(printf 'Authorization: Bearer %s' "$GRAFANA_CLOUD_TOKEN") "$url")" || out=$'\n000'
    else
      out="$(curl "${args[@]}" -H @<(printf 'Authorization: Bearer %s' "$GRAFANA_CLOUD_TOKEN") "$url" </dev/null)" || out=$'\n000'
    fi
  fi
  mem_status="${out##*$'\n'}"
  mem_body="${out%$'\n'*}"
}

# grafana_oncall_base prints the OnCall API's base address from the IRM plugin's settings. It must be https on
# grafana.net, so the token goes nowhere else.
grafana_oncall_base() {
  local url
  grafana_mem GET "$base/api/plugins/grafana-irm-app/settings" bearer
  if ! grafana_ok "$mem_status"; then
    grafana_fail "irm plugin settings failed with HTTP $mem_status: is Grafana IRM installed on the stack?"
    return 1
  fi
  url="$(printf '%s' "$mem_body" | gojq -r '.jsonData.onCallApiUrl // ""' 2>/dev/null)" || url=''
  mem_body=''
  url="${url%/}"
  if [[ ! "$url" =~ ^https://[a-z0-9-]+([.][a-z0-9-]+)*[.]grafana[.]net(/[A-Za-z0-9._/-]*)?$ ]]; then
    grafana_fail "the IRM plugin's onCallApiUrl is not an https address on grafana.net: not used"
    return 1
  fi
  printf '%s/api/v1' "$url"
}

# grafana_integration <oncall base> finds the IRM integration by name, creates it when absent and sets its default
# route's Slack channel when that differs. It leaves the integration's link in irm_link.
irm_link=''
grafana_integration() {
  local oncall="$1" name channel id route
  name="$(gojq -r .integration.name "$dir/notifications.json")"
  channel="$(gojq -r .integration.slackChannelId "$dir/notifications.json")"
  grafana_mem GET "$oncall/integrations/?name=$(printf '%s' "$name" | gojq -sRr @uri)" raw
  grafana_ok "$mem_status" || { grafana_fail "irm integration $name (lookup) failed with HTTP $mem_status"; return 1; }
  id="$(printf '%s' "$mem_body" | gojq -r --arg n "$name" '[.results[]? | select(.name == $n)][0].id // ""')" || id=''
  if [ -z "$id" ]; then
    route="$(gojq -nc --arg n "$name" --arg c "$channel" '{type: "grafana_alerting", name: $n, default_route: {slack: {channel_id: $c}}}')"
    grafana_mem POST "$oncall/integrations/" raw "$route"
    grafana_report "irm integration $name (create)" "$mem_status" || return 1
    irm_link="$(printf '%s' "$mem_body" | gojq -r '.link // ""')" || irm_link=''
  else
    irm_link="$(printf '%s' "$mem_body" | gojq -r --arg n "$name" '[.results[] | select(.name == $n)][0].link // ""')" || irm_link=''
    if printf '%s' "$mem_body" | gojq -e --arg n "$name" --arg c "$channel" '[.results[] | select(.name == $n)][0].default_route.slack | (.channel_id == $c and .enabled == true)' >/dev/null 2>&1; then
      echo "irm integration $name (unchanged): $mem_status"
    else
      route="$(gojq -nc --arg c "$channel" '{default_route: {slack: {channel_id: $c, enabled: true}}}')"
      grafana_mem PUT "$oncall/integrations/$id/" raw "$route"
      grafana_report "irm integration $name (update)" "$mem_status" || return 1
    fi
  fi
  mem_body=''
  if [ -z "$irm_link" ]; then
    grafana_fail "irm integration $name has no link in the OnCall API's answer"
    return 1
  fi
}

# grafana_contact_point upserts the contact point of type oncall at irm_link (PUT by uid, POST on 404), then drops
# the link. The request body, which holds it, goes to curl on stdin and is never printed.
grafana_contact_point() {
  local uid name body
  uid="$(gojq -r .contactPoint.uid "$dir/notifications.json")"
  name="$(gojq -r .contactPoint.name "$dir/notifications.json")"
  body="$(gojq -nc --arg uid "$uid" --arg name "$name" --arg url "$irm_link" '{uid: $uid, name: $name, type: "oncall", settings: {url: $url}}')"
  irm_link=''
  grafana_mem PUT "$base/api/v1/provisioning/contact-points/$uid" bearer "$body"
  if [ "$mem_status" = 404 ]; then
    grafana_mem POST "$base/api/v1/provisioning/contact-points" bearer "$body"
    body=''
    mem_body=''
    grafana_report "contact point $uid (create)" "$mem_status"
  else
    body=''
    mem_body=''
    grafana_report "contact point $uid (update)" "$mem_status"
  fi
}

# grafana_policy puts the folder's route into the notification policy tree: first level, matched by receiver name,
# replaced where it stands or put first when absent. Every other route and the root's settings stay as found.
grafana_policy() {
  local name title status how
  name="$(gojq -r .contactPoint.name "$dir/notifications.json")"
  title="$(gojq -r .title "$dir/folder.json")"
  status="$(grafana_api GET /api/v1/provisioning/policies)"
  grafana_report "notification policies (read)" "$status" || return 1
  cp "$work/response" "$work/policies.json"
  how="$(gojq -r --arg r "$name" 'if ([.routes[]? | select(.receiver == $r)] | length) > 0 then "replace" else "insert" end' "$work/policies.json")" || {
    grafana_fail "the notification policy tree is not readable"
    return 1
  }
  gojq -c --arg r "$name" --arg f "$title" '
    {receiver: $r, object_matchers: [["grafana_folder", "=", $f]], continue: false} as $route
    | .routes = (if ([.routes[]? | select(.receiver == $r)] | length) > 0
        then [.routes[] | if .receiver == $r then $route else . end]
        else [$route] + (.routes // []) end)' "$work/policies.json" >"$work/policies-body.json"
  status="$(grafana_api PUT /api/v1/provisioning/policies "$work/policies-body.json" no-provenance)"
  grafana_report "notification policy route for $name ($how)" "$status"
}

grafana_notifications() {
  local oncall
  oncall="$(grafana_oncall_base)" || return 1
  grafana_integration "$oncall" || return 1
  grafana_contact_point || return 1
  grafana_policy
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
  grafana_notifications
  echo "grafana: applied ${dir#"$root"/} to $base"
}

grafana_main "$@"
