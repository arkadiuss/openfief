---
title: EODHD prices what the core asks for, under the ledger's symbol
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:22.876Z'
id: SCEN-plugins-eodhd-requested-instruments
type: scenario
---
Given no config, when a sync names FAKEW as FAKEWH and the index FAKE20, then FAKEW's closes are delivered as FAKEWH and FAKE20's levels are fetched as FAKE20.INDX; an ISIN on the watch list with symbol SYNTL that the core asks for as SYNT2 is delivered under both; an ambiguous ISIN and an index name with a space are problems on every sync; a request naming an invalid ISIN is InvalidArgument.