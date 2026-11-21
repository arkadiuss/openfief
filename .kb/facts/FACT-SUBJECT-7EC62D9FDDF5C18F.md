---
title: Sync Delivery Transaction Timeout
status: active
fact_kind: subject
subject_key: sync.delivery_transaction_timeout
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-7EC62D9FDDF5C18F
type: fact
---
The single transaction in which a sync's delivery is written, with every statement bounded by a configurable timeout, five minutes by default; past it the delivery fails whole as the class delivery-timeout.