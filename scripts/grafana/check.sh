#!/usr/bin/env bash
# Offline check of deploy/grafana (task grafana:check; docs/dev/demo.md#grafana). It reads the files and calls nothing:
#   - every file is valid JSON, and the folder, each dashboard and each alert rule has a fixed uid (a dashboard's file
#     is named for its uid, a rule's for its uid without the "openfief-" prefix, and uids are unique);
#   - a dashboard carries no database id, and its panels' ids are unique;
#   - a rule is in the folder's uid and in a rule group, and its queries and expressions are on the stack's own
#     datasources: grafanacloud-prom, grafanacloud-logs, grafanacloud-traces, or __expr__ for an expression;
#   - every datasource a dashboard names is one of those, and a metric a query names is one of the five the core
#     exports (sdk/go/telemetry/metrics.go) as Prometheus names it (dots to underscores, _total on a counter, _bucket
#     on a histogram), so a typo or an unlisted metric fails here rather than showing an empty panel;
#   - nothing in the files looks like a secret: a Grafana or Cloud token, a bearer credential, a JSON web token, an
#     address or a password.
# Usage: check.sh [directory]   (default deploy/grafana; gojq on PATH)
set -euo pipefail
shopt -s inherit_errexit

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
dir="$(cd "${1:-$here/../../deploy/grafana}" && pwd)"
failures=0

check_fail() {
  echo "grafana check: $*" >&2
  failures=$((failures + 1))
}

# The names the core exports, as Prometheus stores them.
allowed_metrics='openfief_ui_client_events_total openfief_ui_turnstile_token_seconds_bucket openfief_ui_turnstile_token_seconds_sum openfief_ui_turnstile_token_seconds_count openfief_mcp_unauthenticated_total openfief_authz_availability_changes_total openfief_authz_available'
allowed_datasources='["grafanacloud-prom","grafanacloud-logs","grafanacloud-traces","__expr__"]'

if [ ! -d "$dir" ]; then
  echo "grafana check: no directory $dir" >&2
  exit 1
fi

# check_json fails for a file that is no JSON object.
check_json() {
  if ! gojq -e 'type == "object"' "$1" >/dev/null 2>&1; then
    check_fail "$1 is not a JSON object"
    return 1
  fi
}

check_folder() {
  local file="$dir/folder.json"
  [ -f "$file" ] || { check_fail "no folder.json"; return; }
  check_json "$file" || return 0
  if ! gojq -e '.uid == "openfief-demo" and (.title | type == "string" and length > 0)' "$file" >/dev/null; then
    check_fail "folder.json must have the uid openfief-demo and a title"
  fi
}

check_dashboards() {
  local file uid seen=' ' n=0
  for file in "$dir"/dashboards/*.json; do
    [ -e "$file" ] || break
    n=$((n + 1))
    check_json "$file" || continue
    uid="$(gojq -r '.uid // ""' "$file")"
    if [[ ! "$uid" =~ ^openfief-[a-z0-9-]+$ ]]; then
      check_fail "$file: the uid must be fixed and start with openfief-, got '$uid'"
    elif [ "$(basename "$file" .json)" != "${uid#openfief-}" ]; then
      check_fail "$file: the file must be named for its uid without openfief-: ${uid#openfief-}.json"
    fi
    [[ "$seen" == *" $uid "* ]] && check_fail "$file: the uid $uid is used twice"
    seen="$seen$uid "
    if ! gojq -e '(has("id") | not) or .id == null' "$file" >/dev/null; then
      check_fail "$file: a dashboard must not carry a database id"
    fi
    if ! gojq -e '[.panels[]?.id] | (length > 0 and length == (unique | length))' "$file" >/dev/null; then
      check_fail "$file: panels need unique ids"
    fi
    if ! gojq -e --argjson ok "$allowed_datasources" '[.. | objects | select(has("datasource")) | .datasource | (if type == "object" then .uid else . end)] | all(. as $u | $ok | index($u) != null)' "$file" >/dev/null; then
      check_fail "$file: a datasource is not one of the stack's: grafanacloud-prom, grafanacloud-logs, grafanacloud-traces"
    fi
    check_metrics "$file" '[.. | objects | .expr? | strings]'
  done
  [ "$n" -gt 0 ] || check_fail "no dashboard in $dir/dashboards"
}

# check_metrics <file> <gojq filter that lists the queries>: every openfief_* name in a query is an allowed one.
check_metrics() {
  local name
  while IFS= read -r name; do
    [ -n "$name" ] || continue
    if [[ " $allowed_metrics " != *" $name "* ]]; then
      check_fail "$1: the query names $name, which is none of the metrics the core exports"
    fi
  done < <(gojq -r "$2 | .[]" "$1" | grep -oE 'openfief_[a-z_]+' | sort -u || true)
}

check_rules() {
  local file uid seen=' ' n=0
  for file in "$dir"/alerts/*.json; do
    [ -e "$file" ] || break
    n=$((n + 1))
    check_json "$file" || continue
    uid="$(gojq -r '.uid // ""' "$file")"
    if [[ ! "$uid" =~ ^openfief-[a-z0-9-]+$ ]] || [ "${#uid}" -gt 40 ]; then
      check_fail "$file: the uid must be fixed, start with openfief- and be at most 40 characters, got '$uid'"
    elif [ "$(basename "$file" .json)" != "${uid#openfief-}" ]; then
      check_fail "$file: the file must be named for its uid without openfief-: ${uid#openfief-}.json"
    fi
    [[ "$seen" == *" $uid "* ]] && check_fail "$file: the uid $uid is used twice"
    seen="$seen$uid "
    if ! gojq -e '.folderUID == "openfief-demo" and (.ruleGroup | type == "string" and length > 0) and (.title | type == "string" and length > 0)' "$file" >/dev/null; then
      check_fail "$file: a rule needs a title, a ruleGroup and folderUID openfief-demo"
    fi
    if ! gojq -e '.condition as $c | (.data | map(.refId) | index($c)) != null' "$file" >/dev/null; then
      check_fail "$file: the condition names no query or expression of the rule"
    fi
    if ! gojq -e --argjson ok "$allowed_datasources" '.data | all(.datasourceUid as $u | $ok | index($u) != null)' "$file" >/dev/null; then
      check_fail "$file: a datasourceUid is not one of the stack's: grafanacloud-prom, grafanacloud-logs, grafanacloud-traces, __expr__"
    fi
    check_metrics "$file" '[.data[].model.expr? | strings]'
  done
  [ "$n" -gt 0 ] || check_fail "no alert rule in $dir/alerts"
}

# check_secrets fails when a file holds what only a secret looks like.
check_secrets() {
  local file
  while IFS= read -r file; do
    if grep -qE 'glc_[A-Za-z0-9+/=_-]{8,}|glsa_[A-Za-z0-9_]{8,}|eyJ[A-Za-z0-9_-]{10,}[.][A-Za-z0-9_-]{10,}|[Bb]earer +[A-Za-z0-9._~+/=-]{8,}|[Bb]asic +[A-Za-z0-9+/=]{12,}|"(password|passwd|secret|token|api[_-]?key|authorization)"[[:space:]]*:[[:space:]]*"[^"]+"|[0-9]{1,3}([.][0-9]{1,3}){3}|[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+[.][a-z]{2,}' "$file"; then
      check_fail "$file: looks like it holds a secret, an address or an email (nothing account-specific belongs in git)"
    fi
    if grep -qE '[.]grafana[.]net' "$file"; then
      check_fail "$file: names a grafana.net address; the stack's address is GRAFANA_URL, never in git"
    fi
  done < <(find "$dir" -type f)
}

check_folder
check_dashboards
check_rules
check_secrets

if [ "$failures" -gt 0 ]; then
  echo "grafana check: $failures problem(s) in $dir" >&2
  exit 1
fi
echo "grafana check: $dir is valid"
