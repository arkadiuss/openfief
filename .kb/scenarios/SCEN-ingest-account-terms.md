---
title: The owner gives an account its tax wrapper and fee rule when creating or changing it
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T22:59:00.334Z'
id: SCEN-ingest-account-terms
type: scenario
---
Given an owner session, when the owner creates an account without a wrapper, then it is taxable with the zero fee rule; with IKZE and a fee rule, then both are recorded and revalue the account; changing them records the new ones; a wrapper the pack lacks or a fee rule out of range is refused before anything is stored.