---
title: Demo Grafana Apply
status: active
fact_kind: subject
subject_key: demo.grafana_apply
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T22:41:08.000Z'
id: FACT-SUBJECT-F88D064C852A3514
type: fact
---
The task grafana:apply and its refusals: it puts deploy/grafana, the folder openfief-demo with its dashboards and alert rules, each under a fixed uid, into the demo's Grafana Cloud stack by upsert, and refuses a missing or non-grafana.net https GRAFANA_URL, a missing GRAFANA_CLOUD_TOKEN and files that fail task grafana:check before it calls anything. It does not cover the demo's collector.