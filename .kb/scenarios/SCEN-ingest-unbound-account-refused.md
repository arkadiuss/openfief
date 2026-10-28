---
title: A delivery naming an account the instance is not bound to is refused whole
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:04.524Z'
id: SCEN-ingest-unbound-account-refused
type: scenario
---
Given connector instance xtb-main bound to account acc-1, when its delivery holds a transaction or a balance assertion in acc-2, or the instance has no binding at all, then nothing is imported into the ledger or the valuation stores and the batch fails with an error naming the instance and every unbound account; a delivery of prices only needs no binding.