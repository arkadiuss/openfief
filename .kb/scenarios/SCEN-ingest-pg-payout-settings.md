---
title: The ingest schema keeps an account's payout setting and forgets it with the account
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T11:49:39.692Z'
id: SCEN-ingest-pg-payout-settings
type: scenario
---
Given an account with no stored payout setting, when it is read, then it is the default, as cash; when a setting with earlier settings is saved, then it reads back exactly; when the account is forgotten, then the default reads again.