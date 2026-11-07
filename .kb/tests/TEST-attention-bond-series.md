---
title: A bond series held a week under a provisional ISIN or without terms raises one decision item
status: active
sourceFile: internal/attention/app/bondseries_test.go
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:22.087Z'
id: TEST-attention-bond-series
type: test
---
Runs IncompleteBondSeries on 2026-11-07 over a fake ledger of two accounts and stored terms: series held seven days or more under a provisional ISIN, without terms, or both, are one condition each, subject the ISIN, since the earliest day any account held them, with what is missing; a series held less, one with its registry ISIN and terms, a redeemed one and a listed instrument are none; the kind is a decision and its scope the instrument.