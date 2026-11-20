---
title: Audit Entry Actor
status: active
fact_kind: subject
subject_key: audit.entry_actor
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-93F160F2CA330493
type: fact
distinct_from:
  - audit.admin_actor_entries
---
Who an audit entry names as actor: whoever asked, the owner or a viewer, by session ID, which the log resolves to the person, and never the owner for a viewer's question. It differs from audit.admin_actor_entries, which covers changes made through the core's command line.