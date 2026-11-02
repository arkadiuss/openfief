---
title: Later statements continue an account from its go-live; an opening upload seeds only unstarted accounts
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T13:17:37.196Z'
id: SCEN-ingest-keeps-from-go-live-opening-uploads
type: scenario
---
Given a seeded account, when a sync or an uploaded statement in transactions mode reaches back before its go-live date, then its transactions and ticker entries dated on or before go-live are left out as skipped ingest problems and only what came after is booked or held; and given a statement of several accounts uploaded in opening mode for an account not started yet, when other bound accounts have started, by go-live or from their whole history, then their seeding is left out as skipped problems while the account it was uploaded for is kept whole for the ledger to decide (D-140).