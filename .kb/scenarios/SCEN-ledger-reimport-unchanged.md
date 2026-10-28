---
title: Re-importing the same row changes nothing
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:32:08.771Z'
id: SCEN-ledger-reimport-unchanged
type: scenario
---
Given a buy from source test with institution ID b-1 stored as version 1, when the same row is imported again, even with its postings in another order or 1500.50 written as 1500.500, then the decision is unchanged at version 1 and nothing is appended. Assumes the connector gives the institution's own ID and the same source for every import of that account.