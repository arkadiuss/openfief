---
title: Demo Grafana Notifications
status: active
fact_kind: subject
subject_key: demo.grafana_notifications
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T23:05:19.000Z'
id: FACT-SUBJECT-88B8A72DD624CE0D
type: fact
---
How task grafana:apply provisions the route of the demo's alerts to Slack through Grafana IRM: it finds the integration named in deploy/grafana/notifications.json through the OnCall API, creates it when absent, sets its default route to the Slack channel when that differs, and fails with a clear error and a non-zero exit when the IRM API fails. It does not cover the dashboards and alert rules that task grafana:apply also puts into the stack.