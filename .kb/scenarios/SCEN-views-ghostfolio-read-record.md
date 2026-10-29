---
title: The view reads the record only through the read API
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:31.974Z'
id: SCEN-views-ghostfolio-read-record
type: scenario
---
Given a read API whose account acct-bonds has three transactions served two per page and stored bond valuations through 2025-01-06, when the view syncs, then it reads every account, follows each next page token to the last page, reads the valuation status and the bond valuations from the first bond transaction through the valued-through date in ranges of at most 3700 days, and reads nothing for an account that never held a bond.