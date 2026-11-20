---
title: task grafana:apply is idempotent, keeps the token to the stack and refuses what it must; grafana:check names each fault
status: active
origin:
  kind: agent
  ref: 'Author request: the demo''s Grafana Cloud stack as code, git-tracked JSON plus an idempotent apply script (docs/dev/demo.md#grafana), 2026-11-20'
  recorded_at: '2026-11-20T12:32:31.988Z'
id: SCEN-demo-grafana-as-code
type: scenario
---
Given a fake curl that plays a stack keeping what it is given, when task grafana:apply runs twice, then the first run creates the folder and each alert rule (a 404 on the PUT, then a POST) and applies the dashboard with no database id and overwrite, the second updates them in place with no create and leaves the stack's state as it was, each call carries the token as a bearer credential and rules are sent with X-Disable-Provenance, and the output holds the names and statuses and neither the token nor a response body; a missing GRAFANA_URL or token, an http address, another host, a look-alike host, credentials, a path, a port or a query in the address, a token the stack refuses, a failing call and an unreachable stack each fail with the status and no token in the output, and a refusal before any call makes none; and task grafana:check passes the committed files and refuses a missing uid, a database id, a datasource or metric outside the stack's and the core's, a rule outside the folder, a condition that names nothing, a uid used twice, a file not named for its uid, a file that is no JSON and a token, bearer credential, address or grafana.net name in any file.