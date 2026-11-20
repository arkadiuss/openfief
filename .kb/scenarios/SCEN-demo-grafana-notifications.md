---
title: task grafana:apply routes alerts to Slack through IRM idempotently, preserves the other routes and never shows the integration's link
status: active
origin:
  kind: agent
  ref: 'Author request: route the demo''s Grafana alerts to Slack #openfief-demo through Grafana IRM, provisioned by task grafana:apply (docs/dev/demo.md#grafana), 2026-11-20'
  recorded_at: '2026-11-20T22:31:03.989Z'
id: SCEN-demo-grafana-notifications
type: scenario
---
Given a fake curl that plays a stack and its OnCall API, when task grafana:apply runs, then the first run creates the IRM integration as grafana_alerting with the channel in git as its default route, creates a contact point of type oncall holding the integration's link and puts one first-level route for the folder first in the policy tree with every other route and the default receiver as found; the second run creates nothing and leaves the stack's state as it was; an integration on another channel gets its route set; a stale route for the receiver is replaced in place; the OnCall calls carry the token alone and the stack's address, and the link and the token are in no output or log; a failing IRM API or missing plugin settings fail the apply with the status and no body; and task grafana:check refuses a notifications file with a channel name or lower-case ID, a URL, or a contact point without a fixed uid.