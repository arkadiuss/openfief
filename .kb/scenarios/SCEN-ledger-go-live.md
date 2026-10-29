---
title: An account's go-live is its latest seeding transaction's date
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:16.122Z'
id: SCEN-ledger-go-live
type: scenario
---
Given an account with an opening lot dated 2026-03-31 and opening cash dated 2026-03-30, its go-live date is 2026-03-31; with deposits and buys only, or an opening lot moved to another account by a correction, it is not seeded; an invalid log is an error, and an unknown account is ErrUnknownAccount.