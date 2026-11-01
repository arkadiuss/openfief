---
title: Later statements continue an account from its go-live
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:34.549Z'
id: SCEN-ingest-keeps-from-go-live-uploads
type: scenario
---
Given a seeded account, when a sync or an uploaded statement in transactions mode reaches back before its go-live date, then its transactions and ticker entries dated on or before go-live are left out as skipped ingest problems and only what came after is booked or held, while an opening upload is kept whole for the ledger to decide.