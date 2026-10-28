---
title: The demo history reconciles and its statements map cleanly
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:41.636Z'
id: SCEN-demo-consistent-history
type: scenario
---
Given the generated dataset, when the IBKR statements are mapped by the ibkrflex connector and the XTB and bond batches are taken as they are, then the opening statement has no ingest problem, the monthly statements only the intended skipped SUMMARY dividend rows, every balance assertion matches the derived balance, and the COI0128 coupon paid on 2026-01-10 is 30 bonds at 5.95 PLN.