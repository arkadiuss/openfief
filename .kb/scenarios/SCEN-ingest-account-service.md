---
title: The owner creates an account from a broker account the connector reported
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:36.651Z'
id: SCEN-ingest-account-service
type: scenario
---
Given ibkr-main reported the synthetic broker accounts U0000001 and U0000002, when the owner creates the account IBKR main in satellite for U0000001, then the core stores ibkr-main at ibkr, binds ibkr-main to it and maps U0000001 to it, and the next sync books U0000001's deposit there; creating another account for U0000001, or one for a broker account never reported, is refused before anything is stored.