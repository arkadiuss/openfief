---
title: Front Door Client Address
status: active
fact_kind: subject
subject_key: front_door.client_address
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-6EE8B238C4BB1FA1
type: fact
---
How the core's front door finds the browser's address: from the connection, or from the header of a configured reverse proxy only when the connection comes from a proxy it is told to trust.