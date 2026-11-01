---
title: The owner pays out a hand-entered bond account from today
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:19.120Z'
id: SCEN-ingest-bond-payout-setting
type: scenario
---
Given the owner's session, when the owner changes a treasury-bonds account to paid out through AccountService, then the account is stored paid out since today and answered so; asking to pay out an account at another institution, or a setting outside the declared ones, fails with INVALID_ARGUMENT and stores nothing.