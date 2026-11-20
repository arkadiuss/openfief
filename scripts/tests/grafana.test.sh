#!/usr/bin/env bash
# Tests for scripts/grafana/apply.sh and check.sh (task grafana:apply and grafana:check) against a fake curl on PATH:
# no network, no real stack, no real token (kibi: TEST-tooling-grafana-apply). The fake curl plays a stack that keeps
# what it is given, so a second apply meets what the first made: 404 for what it doesn't hold, 200 or 201 for what it
# does. It logs every call's method and address, reads the Authorization header out of the header file curl is given
# and says only whether the token is the right one, and never logs the header, so the checks can grep for the token.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
failures=0
cases=0
# A synthetic value that looks like nothing real; every check that must not show it greps for it.
fake_token='glsa_FAKE0123456789abcdef_for_tests_only'
stack=https://demo-stack.grafana.net
# The IRM integration's link is a credential: a synthetic one, which must be in no output or log, and the OnCall API's address.
fake_link='https://oncall-fake.grafana.net/integrations/v1/grafana_alerting/FAKELINKSECRET123/'
oncall_base=https://oncall-fake.grafana.net/oncall

mkdir -p "$work/bin" "$work/state"
cat >"$work/bin/curl" <<'EOF'
#!/usr/bin/env bash
# A fake Grafana. Arguments it understands: -sS, -o file, -w format, -X method, --max-time n, -H header or @file,
# --data-binary @file, and the address last.
method=GET out='' body='' bearer=wrong rawtoken=wrong grafanaurl=wrong provenance=no wfmt=''
while [ $# -gt 1 ]; do
  case "$1" in
    -o) out="$2"; shift 2 ;;
    -w) wfmt="$2"; shift 2 ;;
    --max-time) shift 2 ;;
    -X) method="$2"; shift 2 ;;
    --data-binary)
      if [ "$2" = @- ]; then cat >"$FAKE_STATE/.stdin"; body="$FAKE_STATE/.stdin"; else body="${2#@}"; fi
      shift 2
      ;;
    -H)
      case "$2" in
        @*)
          # The header file is a pipe: read it once.
          header="$(cat "${2#@}")"
          [ "$header" = "Authorization: Bearer $FAKE_TOKEN" ] && bearer=right
          [ "$header" = "Authorization: $FAKE_TOKEN" ] && rawtoken=right
          ;;
        "X-Grafana-URL: $FAKE_BASE") grafanaurl=right ;;
        X-Disable-Provenance:*) provenance=yes ;;
      esac
      shift 2
      ;;
    *) shift ;;
  esac
done
url="$1"
oncall=no auth=wrong
case "$url" in
  "$FAKE_ONCALL"/*) oncall=yes; path="ONCALL${url#"$FAKE_ONCALL"}"
    if [ "$rawtoken" = right ] && [ "$grafanaurl" = right ]; then auth=right; fi ;;
  *) path="${url#"$FAKE_BASE"}"
    if [ "$bearer" = right ]; then auth=right; fi ;;
esac
echo "$method $path auth=$auth provenance=$provenance" >>"$FAKE_LOG"
# reply <status> [<body>]: with -o the body goes to the file, else to stdout before "\n<status>" as -w asks.
reply() {
  local b="${2:-}"
  [ -n "$b" ] || b='{"message":"fake"}'
  if [ -n "$out" ]; then
    printf '%s\n' "$b" >"$out"; printf '%s' "$1"
  else
    printf '%s' "$b"; [ -n "$wfmt" ] && printf '\n%s' "$1"
  fi
  exit 0
}
[ "$auth" = right ] || reply 401
if [ -n "${FAKE_FAIL_PATH:-}" ] && [ "$path" = "$FAKE_FAIL_PATH" ]; then reply 500; fi
case "$method $path" in
  "GET /api/plugins/grafana-irm-app/settings") reply 200 "{\"jsonData\":{\"onCallApiUrl\":\"$FAKE_ONCALL\"}}" ;;
  "GET ONCALL/api/v1/integrations/?name="*)
    if [ -e "$FAKE_STATE/integration" ]; then reply 200 "{\"results\":[$(cat "$FAKE_STATE/integration")]}"; fi
    reply 200 '{"results":[]}'
    ;;
  "POST ONCALL/api/v1/integrations/")
    gojq -c '. + {id: "INTEGRATION1", link: "'"$FAKE_LINK"'", default_route: {slack: {channel_id: .default_route.slack.channel_id, enabled: true}}}' "$body" >"$FAKE_STATE/integration"
    reply 201 "$(cat "$FAKE_STATE/integration")"
    ;;
  "PUT ONCALL/api/v1/integrations/INTEGRATION1/")
    gojq -c --slurpfile new "$body" '.default_route = $new[0].default_route' "$FAKE_STATE/integration" >"$FAKE_STATE/integration.new"
    mv "$FAKE_STATE/integration.new" "$FAKE_STATE/integration"
    reply 200 "$(cat "$FAKE_STATE/integration")"
    ;;
  "PUT /api/v1/provisioning/contact-points/"*)
    if [ -e "$FAKE_STATE/contact-point" ]; then cp "$body" "$FAKE_STATE/contact-point"; reply 202 "$(cat "$body")"; fi
    reply 404
    ;;
  "POST /api/v1/provisioning/contact-points") cp "$body" "$FAKE_STATE/contact-point"; reply 202 "$(cat "$body")" ;;
  "GET /api/v1/provisioning/policies")
    if [ -e "$FAKE_STATE/policies" ]; then reply 200 "$(cat "$FAKE_STATE/policies")"; fi
    reply 200 '{"receiver":"default-email","group_by":["grafana_folder","alertname"],"routes":[{"receiver":"other","object_matchers":[["team","=","x"]],"continue":true},{"receiver":"third","matchers":["a=b"]}]}'
    ;;
  "PUT /api/v1/provisioning/policies") cp "$body" "$FAKE_STATE/policies"; reply 202 ;;
  "GET /api/folders/"*) [ -e "$FAKE_STATE/folder" ] && reply 200 || reply 404 ;;
  "POST /api/folders") : >"$FAKE_STATE/folder"; reply 200 ;;
  "PUT /api/folders/"*) reply 200 ;;
  "POST /api/dashboards/db") cp "$body" "$FAKE_STATE/dashboard-$(gojq -r .dashboard.uid "$body")"; reply 200 ;;
  "PUT /api/v1/provisioning/alert-rules/"*)
    uid="${path##*/}"
    if [ -e "$FAKE_STATE/rule-$uid" ]; then cp "$body" "$FAKE_STATE/rule-$uid"; reply 200; fi
    reply 404
    ;;
  "POST /api/v1/provisioning/alert-rules")
    uid="$(gojq -r .uid "$body")"
    cp "$body" "$FAKE_STATE/rule-$uid"
    reply 201
    ;;
  *) echo "fake curl: unexpected call: $method $path " >&2; reply 599 ;;
esac
EOF
chmod +x "$work/bin/curl"

# run_apply runs apply.sh with the fake curl first on PATH, the stack's address and the token, plus the assignments
# given; "unset NAME" arguments are not supported, so a test that needs one passes it empty.
run_apply() {
  : >"$work/log"
  env -i PATH="$work/bin:$PATH" HOME="$work" FAKE_LOG="$work/log" FAKE_STATE="$work/state" FAKE_BASE="$stack" \
    FAKE_TOKEN="$fake_token" FAKE_ONCALL="$oncall_base" FAKE_LINK="$fake_link" GRAFANA_URL="$stack" GRAFANA_CLOUD_TOKEN="$fake_token" "$@" \
    bash "$root/scripts/grafana/apply.sh" >"$work/out" 2>&1
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

# nowhere reports that the value is in none of the run's outputs: the fake's log and the script's own output.
nowhere() {
  ! grep -rqF -- "$1" "$work/log" "$work/out"
}

# refuses reports that a run fails before it calls anything, and says why.
refuses() {
  local name="$1" want="$2"
  shift 2
  if run_apply "$@"; then r=no; else r=yes; fi
  check "$name: fails" "$r"
  check "$name: says why" "$(yes_if grep -qF -- "$want" "$work/out")"
  check "$name: calls nothing" "$(yes_if test ! -s "$work/log")"
  check "$name: no token in its output" "$(yes_if nowhere "$fake_token")"
}

# What the script must refuse: a missing or wrong address, a missing token.
refuses "no GRAFANA_URL" "GRAFANA_URL is not set" GRAFANA_URL=
refuses "an http address" "must be an https address on grafana.net" GRAFANA_URL=http://demo-stack.grafana.net
refuses "another host" "must be an https address on grafana.net" GRAFANA_URL=https://grafana.example.com
refuses "a host that only ends like grafana.net" "must be an https address on grafana.net" GRAFANA_URL=https://demo-stack.grafana.net.example.com
refuses "a host that only starts like it" "must be an https address on grafana.net" GRAFANA_URL=https://grafana.net.example.com
refuses "a look-alike without the dot" "must be an https address on grafana.net" GRAFANA_URL=https://demo-stackgrafana.net
refuses "credentials in the address" "must be an https address on grafana.net" GRAFANA_URL=https://user:pw@demo-stack.grafana.net
refuses "a path in the address" "must be an https address on grafana.net" GRAFANA_URL=https://demo-stack.grafana.net/grafana
refuses "a port in the address" "must be an https address on grafana.net" GRAFANA_URL=https://demo-stack.grafana.net:8443
refuses "a query in the address" "must be an https address on grafana.net" 'GRAFANA_URL=https://demo-stack.grafana.net?x=1'
refuses "no token" "GRAFANA_CLOUD_TOKEN is not set" GRAFANA_CLOUD_TOKEN=

# The first apply makes everything: the folder, the dashboard and each rule (404 on the PUT, then a POST).
rm -rf "$work/state" && mkdir -p "$work/state"
if run_apply; then r=yes; else r=no; fi
check "the first apply succeeds" "$r"
cp "$work/out" "$work/out-first"
check "it creates the folder" "$(yes_if grep -qxF 'folder openfief-demo (create): 200' "$work/out")"
check "it applies the dashboard" "$(yes_if grep -qxF 'dashboard openfief-demo: 200' "$work/out")"
rules="$(find "$root/deploy/grafana/alerts" -name '*.json' | wc -l)"
check "there are alert rules to apply" "$(yes_if test "$rules" -ge 3)"
for f in "$root"/deploy/grafana/alerts/*.json; do
  uid="$(gojq -r .uid "$f")"
  check "it creates the rule $uid" "$(yes_if grep -qxF "alert rule $uid (create): 201" "$work/out")"
  check "the rule $uid reached the stack as it is in git" "$(yes_if cmp -s "$f" "$work/state/rule-$uid")"
done
check "the dashboard reached the stack with no database id, in the folder, overwriting" \
  "$(yes_if gojq -e '.dashboard.id == null and .folderUid == "openfief-demo" and .overwrite == true and .dashboard.uid == "openfief-demo"' "$work/state/dashboard-openfief-demo")"
check "every call carried the token as the stack's bearer credential" "$(yes_if test "$(grep -c 'auth=right' "$work/log")" -eq "$(wc -l <"$work/log")")"
check "every call went to the stack's own address or the OnCall API the plugin names, and nowhere else" \
  "$(yes_if test "$(grep -c '^[A-Z]* \(/api/\|ONCALL/api/v1/\)' "$work/log")" -eq "$(wc -l <"$work/log")")"
check "rules are sent without provenance, so they stay editable" \
  "$(yes_if test "$(grep -c '^\(PUT\|POST\) /api/v1/provisioning/alert-rules.* provenance=yes$' "$work/log")" -eq "$((rules * 2))")"
check "the folder and the dashboard are not sent with it" "$(yes_if grep -qx 'POST /api/dashboards/db auth=right provenance=no' "$work/log")"
check "the token is not in the output" "$(yes_if nowhere "$fake_token")"
check "the output holds no response body" "$(yes_if nowhere '"message"')"
check "it creates the IRM integration" "$(yes_if grep -qxF 'irm integration OpenFief demo (create): 201' "$work/out")"
check "the integration was made as grafana_alerting, routed to the channel in git" \
  "$(yes_if gojq -e --arg c "$(gojq -r .integration.slackChannelId "$root/deploy/grafana/notifications.json")" '.type == "grafana_alerting" and .name == "OpenFief demo" and .default_route.slack.channel_id == $c' "$work/state/integration")"
check "it creates the contact point" "$(yes_if grep -qxF 'contact point openfief-demo-irm (create): 202' "$work/out")"
check "the contact point is of type oncall and holds the integration's link" \
  "$(yes_if gojq -e --arg l "$fake_link" '.uid == "openfief-demo-irm" and .name == "OpenFief demo (IRM → Slack)" and .type == "oncall" and .settings.url == $l' "$work/state/contact-point")"
check "the policy tree has the one route, first, with the folder matcher and no continue" \
  "$(yes_if gojq -e '[.routes[] | select(.receiver == "OpenFief demo (IRM → Slack)")] == [{receiver: "OpenFief demo (IRM → Slack)", object_matchers: [["grafana_folder", "=", "OpenFief demo"]], continue: false}] and .routes[0].receiver == "OpenFief demo (IRM → Slack)"' "$work/state/policies")"
check "the other routes and the default receiver are exactly as found" \
  "$(yes_if gojq -e '.receiver == "default-email" and .group_by == ["grafana_folder","alertname"] and .routes[1:] == [{"receiver":"other","object_matchers":[["team","=","x"]],"continue":true},{"receiver":"third","matchers":["a=b"]}]' "$work/state/policies")"
check "the OnCall calls used the token alone and the stack's address; the stack's calls the bearer form" "$(yes_if test "$(grep -c 'auth=right' "$work/log")" -eq "$(wc -l <"$work/log")")"
check "the policy and the contact point are sent without provenance" \
  "$(yes_if grep -qx 'PUT /api/v1/provisioning/policies auth=right provenance=yes' "$work/log")"
check "the integration's link is in no output or log" "$(yes_if nowhere "$fake_link")"
check "nor is any part of it" "$(yes_if nowhere FAKELINKSECRET123)"

# A second apply finds everything and changes it in place: PUTs, no POST of a rule or the folder, the same state.
cp -r "$work/state" "$work/state-first"
if run_apply; then r=yes; else r=no; fi
check "the second apply succeeds" "$r"
check "it updates the folder" "$(yes_if grep -qxF 'folder openfief-demo (update): 200' "$work/out")"
check "it applies the dashboard again" "$(yes_if grep -qxF 'dashboard openfief-demo: 200' "$work/out")"
for f in "$root"/deploy/grafana/alerts/*.json; do
  uid="$(gojq -r .uid "$f")"
  check "it updates the rule $uid" "$(yes_if grep -qxF "alert rule $uid (update): 200" "$work/out")"
done
check "it creates nothing: no POST of a folder, a rule or a contact point" "$(yes_if test "$(grep -c '^POST /api/\(folders\|v1\)' "$work/log")" -eq 0)"
check "it finds the integration and leaves it alone: no POST or PUT to the OnCall API" "$(yes_if test "$(grep -c '^\(POST\|PUT\) ONCALL' "$work/log")" -eq 0)"
check "it reports the integration unchanged" "$(yes_if grep -qxF 'irm integration OpenFief demo (unchanged): 200' "$work/out")"
check "it replaces the policy route where it stands" "$(yes_if grep -qxF 'notification policy route for OpenFief demo (IRM → Slack) (replace): 202' "$work/out")"
check "the policy tree has still exactly one route for the receiver" \
  "$(yes_if gojq -e '[.routes[] | select(.receiver == "OpenFief demo (IRM → Slack)")] | length == 1' "$work/state/policies")"
check "the integration's link is in no output or log of the second run" "$(yes_if nowhere "$fake_link")"
check "the stack's state is the same after the second apply" "$(yes_if diff -r "$work/state" "$work/state-first")"
check "the second apply prints the same lines but for create and update" \
  "$(yes_if test "$(sed -e 's/ (create)//' -e 's/ (update)//' -e 's/ (insert)//' -e 's/: 201$/: 200/' "$work/out-first")" = "$(sed -e 's/ (create)//' -e 's/ (update)//' -e 's/ (replace)//' -e 's/ (unchanged)//' "$work/out")")"
check "the token is not in the output of the second run either" "$(yes_if nowhere "$fake_token")"

# An integration routed to another channel gets its default route set, and nothing else is created.
gojq -c '.default_route.slack.channel_id = "C0OTHERCHAN1"' "$work/state/integration" >"$work/state/integration.new" && mv "$work/state/integration.new" "$work/state/integration"
if run_apply; then r=yes; else r=no; fi
check "an apply over an integration on another channel succeeds" "$r"
check "it updates the integration's route" "$(yes_if grep -qxF 'irm integration OpenFief demo (update): 200' "$work/out")"
check "the integration is on the channel in git again" \
  "$(yes_if gojq -e --arg c "$(gojq -r .integration.slackChannelId "$root/deploy/grafana/notifications.json")" '.default_route.slack.channel_id == $c and .id == "INTEGRATION1"' "$work/state/integration")"
check "it created no second integration" "$(yes_if test "$(grep -c '^POST ONCALL' "$work/log")" -eq 0)"
check "the link is in no output or log after an update" "$(yes_if nowhere "$fake_link")"

# A stale route for the receiver is replaced, not doubled; routes of others stay where they are.
echo '{"receiver":"default-email","routes":[{"receiver":"other"},{"receiver":"OpenFief demo (IRM → Slack)","matchers":["stale=1"],"continue":true},{"receiver":"third"}]}' >"$work/state/policies"
if run_apply; then r=yes; else r=no; fi
check "an apply over a stale route succeeds" "$r"
check "the stale route became the folder's route in place, and the others stayed" \
  "$(yes_if gojq -e '.routes == [{"receiver":"other"},{receiver: "OpenFief demo (IRM → Slack)", object_matchers: [["grafana_folder","=","OpenFief demo"]], continue: false},{"receiver":"third"}] and .receiver == "default-email"' "$work/state/policies")"

# The IRM API failing is a clear error and a failed apply, with neither the link nor the token in the output.
if run_apply FAKE_FAIL_PATH='ONCALL/api/v1/integrations/?name=OpenFief%20demo'; then r=no; else r=yes; fi
check "a failing IRM API fails the apply" "$r"
check "it names the integration and the status" "$(yes_if grep -qF 'irm integration OpenFief demo (lookup) failed with HTTP 500' "$work/out")"
check "it makes no contact point or policy change after it" "$(yes_if test "$(grep -c 'contact-points\|policies' "$work/log")" -eq 0)"
check "the failure shows neither the link nor the token" "$(yes_if nowhere "$fake_link")"
check "and no response body" "$(yes_if nowhere '"results"')"
if run_apply FAKE_FAIL_PATH=/api/plugins/grafana-irm-app/settings; then r=no; else r=yes; fi
check "missing IRM plugin settings fail the apply" "$r"
check "it says so" "$(yes_if grep -qF 'irm plugin settings failed with HTTP 500' "$work/out")"
check "it calls the OnCall API not at all" "$(yes_if test "$(grep -c ONCALL "$work/log")" -eq 0)"

# A wrong token is refused by the stack (401): the script fails and names the status, not the token.
if run_apply GRAFANA_CLOUD_TOKEN=glsa_WRONG0123456789_for_tests_only; then r=no; else r=yes; fi
check "a token the stack refuses fails the apply" "$r"
check "it says HTTP 401" "$(yes_if grep -qF 'failed with HTTP 401' "$work/out")"
check "neither token is in the output" "$(yes_if nowhere glsa_WRONG0123456789_for_tests_only)"
check "the right token is in no output" "$(yes_if nowhere "$fake_token")"

# A failing call fails the apply, names what failed and its status, and prints no body.
if run_apply FAKE_FAIL_PATH=/api/dashboards/db; then r=no; else r=yes; fi
check "a 500 fails the apply" "$r"
check "it names the dashboard and the status" "$(yes_if grep -qF 'dashboard openfief-demo: 500' "$work/out")"
check "it applies nothing after the failure" "$(yes_if test "$(grep -c 'alert-rules' "$work/log")" -eq 0)"
check "the failure prints no response body" "$(yes_if nowhere '"message"')"

# curl itself failing (no route, a refused connection) is status 000 and a failed apply.
cat >"$work/bin/curl-down" <<'EOF'
#!/usr/bin/env bash
echo "curl: (6) Could not resolve host" >&2
exit 6
EOF
chmod +x "$work/bin/curl-down"
mkdir -p "$work/down" && cp "$work/bin/curl-down" "$work/down/curl"
: >"$work/log"
if env -i PATH="$work/down:$PATH" HOME="$work" GRAFANA_URL="$stack" GRAFANA_CLOUD_TOKEN="$fake_token" \
  bash "$root/scripts/grafana/apply.sh" >"$work/out" 2>&1; then r=no; else r=yes; fi
check "an unreachable stack fails the apply" "$r"
check "it says HTTP 000" "$(yes_if grep -qF 'HTTP 000' "$work/out")"
check "the token is not in the output when curl fails" "$(yes_if nowhere "$fake_token")"

# The offline check: the real files pass, and each fault is named. It reads a copy, so the files stay as they are.
run_check() {
  env -i PATH="$PATH" HOME="$work" bash "$root/scripts/grafana/check.sh" "$1" >"$work/out" 2>&1
}
if run_check "$root/deploy/grafana"; then r=yes; else r=no; fi
check "the committed deploy/grafana passes the check" "$r"

# faulty <name> <want> <gojq or sed edit>: copy the tree, break one file, and expect the check to name the fault.
faulty() {
  local name="$1" want="$2" file="$3" filter="$4"
  rm -rf "$work/tree" && cp -r "$root/deploy/grafana" "$work/tree"
  gojq "$filter" "$work/tree/$file" >"$work/tree/$file.new" && mv "$work/tree/$file.new" "$work/tree/$file"
  if run_check "$work/tree"; then r=no; else r=yes; fi
  check "the check refuses $name" "$r"
  check "the check says why: $want" "$(yes_if grep -qF -- "$want" "$work/out")"
}
faulty "a dashboard without a uid" "the uid must be fixed" dashboards/demo.json 'del(.uid)'
faulty "a dashboard with a database id" "must not carry a database id" dashboards/demo.json '.id = 7'
faulty "a dashboard on another datasource" "a datasource is not one of the stack's" dashboards/demo.json '.panels[0].datasource.uid = "some-other-datasource"'
faulty "a query on a metric the core does not export" "which is none of the metrics the core exports" dashboards/demo.json '.panels[4].targets[0].expr = "sum(rate(openfief_ui_client_event_total[5m]))"'
faulty "a rule without a uid" "the uid must be fixed" alerts/telemetry-silent.json 'del(.uid)'
faulty "a rule outside the folder" "folderUID openfief-demo" alerts/telemetry-silent.json '.folderUID = "elsewhere"'
faulty "a rule on another datasource" "a datasourceUid is not one of the stack's" alerts/core-error-logs.json '.data[0].datasourceUid = "grafanacloud-usage"'
faulty "a rule whose condition names nothing" "the condition names no query" alerts/core-error-logs.json '.condition = "Z"'
faulty "a Slack channel name where the ID goes" "integration.slackChannelId must be a Slack channel ID" notifications.json '.integration.slackChannelId = "#all-channel"'
faulty "a lower-case Slack channel ID" "integration.slackChannelId must be a Slack channel ID" notifications.json '.integration.slackChannelId = "c0ckvmndtk6"'
faulty "a URL in the notifications" "holds a URL" notifications.json '.integration.link = "https://oncall.example/integrations/v1/x/"'
faulty "a contact point without a fixed uid" "contactPoint.uid must be fixed" notifications.json '.contactPoint.uid = "demo"'
faulty "a folder with another uid" "uid openfief-demo" folder.json '.uid = "somewhere"'
faulty "a token in an annotation" "looks like it holds a secret" alerts/turnstile-errors.json ".annotations.description = \"use $fake_token\""
faulty "a bearer credential in a description" "looks like it holds a secret" dashboards/demo.json '.description = "Authorization: Bearer abcdefghijklmnop12345"'
faulty "an address in a description" "looks like it holds a secret" dashboards/demo.json '.description = "from 198.51.100.7"'
faulty "the stack's address" "names a grafana.net address" dashboards/demo.json '.description = "see https://stack.grafana.net/d/x"'
rm -rf "$work/tree" && cp -r "$root/deploy/grafana" "$work/tree" && echo '{broken' >"$work/tree/alerts/broken.json"
if run_check "$work/tree"; then r=no; else r=yes; fi
check "the check refuses a file that is no JSON" "$r"
rm -rf "$work/tree" && cp -r "$root/deploy/grafana" "$work/tree" && mv "$work/tree/alerts/telemetry-silent.json" "$work/tree/alerts/other.json"
if run_check "$work/tree"; then r=no; else r=yes; fi
check "the check refuses a rule file not named for its uid" "$r"
rm -rf "$work/tree" && cp -r "$root/deploy/grafana" "$work/tree" && cp "$work/tree/alerts/telemetry-silent.json" "$work/tree/alerts/silent-twin.json"
if run_check "$work/tree"; then r=no; else r=yes; fi
check "the check refuses a uid used twice" "$r"
check "it says so" "$(yes_if grep -qF 'used twice' "$work/out")"

# apply.sh runs the check first: a faulty tree is applied not at all.
rm -rf "$work/state" && mkdir -p "$work/state"
rm -rf "$work/tree" && cp -r "$root/deploy/grafana" "$work/tree" && gojq 'del(.uid)' "$work/tree/dashboards/demo.json" >"$work/tree/d.json" && mv "$work/tree/d.json" "$work/tree/dashboards/demo.json"
if run_apply GRAFANA_DIR="$work/tree"; then r=no; else r=yes; fi
check "apply refuses a tree that fails the check" "$r"
check "and calls nothing" "$(yes_if test ! -s "$work/log")"

echo "grafana.test.sh: $cases checks, $failures failed"
[ "$failures" -eq 0 ]
