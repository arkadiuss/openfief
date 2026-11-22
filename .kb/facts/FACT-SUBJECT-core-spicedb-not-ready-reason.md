---
title: Core SpiceDB Not Ready Reason
status: active
fact_kind: subject
subject_key: core.spicedb_not_ready_reason
origin:
  kind: agent
  recorded_at: '2026-11-22T08:39:14.909Z'
id: FACT-SUBJECT-core-spicedb-not-ready-reason
type: fact
---
What a core that runs SpiceDB says, in its /readyz answer and its log, while it is not ready: which of SpiceDB's prerequisites is missing (the Secret with the preshared key, SpiceDB's role or database in the openfief-db cluster, a key SpiceDB accepts, or SpiceDB answering), by name. It is the wording of the wait, not the readiness gate itself, which the start's reconcile decides.