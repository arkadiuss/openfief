---
title: Syncing the same Flex statement and uploading the same XTB file or bond lot again adds nothing
status: active
tags:
  - acceptance:p10-3
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:35.735Z'
id: SCEN-e2e-reimport-idempotent
type: scenario
---
Given the IBKR account synced from the week statement, the XTB account holding a later export and the bond account holding a lot, when the same statement is synced again, the same export uploaded again and the same lot entered again, then core check ledger lists the same transactions with the same versions for each account as before. Runs on both kind and k3d.