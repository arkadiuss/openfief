---
title: Telemetry Private Metrics Allow List
status: active
fact_kind: subject
subject_key: telemetry.private_metrics_allow_list
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-143E9AFB0E9B8E68
type: fact
---
The allow-list through which metrics leave a process, only over OTLP/HTTP to a configured endpoint, off unless set, so nothing private is exported.