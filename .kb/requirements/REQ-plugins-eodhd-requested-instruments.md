---
title: The EODHD plugin must price each instrument a sync request names, delivering its closes under the symbol the request gives as well as a watched entry's, and must fetch each requested market index as the ticker of its own name plus .INDX unless its config names the ticker, reporting a name that can't be a ticker as an ingest problem.
status: open
text_ref: docs/DESIGN.md#D-93
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-93
  - lane:strict
semantic_text: The EODHD plugin must price each instrument a sync request names, delivering its closes under the symbol the request gives as well as a watched entry's, and must fetch each requested market index as the ticker of its own name plus .INDX unless its config names the ticker, reporting a name that can't be a ticker as an ingest problem.
logic_claims:
  - CLAIM-E76BDF40795B8125
semantic_clauses:
  - The EODHD plugin must price each instrument a sync request names, delivering its closes under the symbol the request gives as well as a watched entry's, and must fetch each requested market index as the ticker of its own name plus .INDX unless its config names the ticker, reporting a name that can't be a ticker as an ingest problem.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: eaff63402a9d54d6f3e5999d2ec750746190a90ba6fbf234e068ed4f9dfd3e97
semantic_inventory:
  - claim_key: CLAIM-E76BDF40795B8125
    claim_text: The EODHD plugin must price each instrument a sync request names, delivering its closes under the symbol the request gives as well as a watched entry's, and must fetch each requested market index as the ticker of its own name plus .INDX unless its config names the ticker, reporting a name that can't be a ticker as an ingest problem
    role: exception
    status: modeled
    span:
      start: 0
      end: 333
origin:
  kind: agent
  recorded_at: '2026-10-29T16:58:12.500Z'
id: REQ-plugins-eodhd-requested-instruments
type: req
---
The EODHD plugin must price each instrument a sync request names, delivering its closes under the symbol the request gives as well as a watched entry's, and must fetch each requested market index as the ticker of its own name plus .INDX unless its config names the ticker, reporting a name that can't be a ticker as an ingest problem.

## Context

D-93 turns the EODHD config's instrument list into an optional watch list: the core names what to price, as the ledger knows it, so prices key to the ledger's instrument (D-35). D-96 identifies an index by the owner's name; deriving the EODHD ticker from it is the narrowest reading. Modeled by Claude Code for the market-data work package (D-93, D-94, D-95, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-93]** Each sync request to a market-data connector carries the instruments the owner holds and the model portfolio's targets, so a newly bought instrument is priced without configuration. An optional list in the instance's config adds instruments to watch.

docs/DESIGN.md#D-93

> **[D-96]** The EODHD connector identifies a market index by the name the owner gives it, such as WIG20.

docs/DESIGN.md#D-96
