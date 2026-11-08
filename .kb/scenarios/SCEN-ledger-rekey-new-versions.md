---
title: The ledger re-keys by new versions under the same keys
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:30.105Z'
id: SCEN-ledger-rekey-new-versions
type: scenario
---
Given transactions in several accounts post to a provisional instrument, when the ledger re-keys it to the registry's instrument, then each gets one new version under its key with only its holding postings moved, in one database transaction whose hooks see the first day re-keyed, display names carry over, and re-keying again changes nothing.