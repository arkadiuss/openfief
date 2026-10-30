---
title: The owner deletes a bond lot as a correction
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T06:33:11.713Z'
id: SCEN-ingest-bond-lot-delete-entry
type: scenario
---
Given a lot entered in a bond account, when the owner deletes it by its series and purchase date, then the core hands the key to the bound connector, which delivers a deletion of each transaction the lot was booked as, keyed and dated as adding it; a key without a date, a non-owner or a lot never entered is refused.