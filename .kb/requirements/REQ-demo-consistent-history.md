---
title: 'The demo history must reconcile: every balance assertion it carries must match the balance derived from its transactions, its opening IBKR statement must map with no ingest problem and each monthly one with only the intended skipped summary rows, and bond coupons must follow the series'' published rates.'
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: 'The demo history must reconcile: every balance assertion it carries must match the balance derived from its transactions, its opening IBKR statement must map with no ingest problem and each monthly one with only the intended skipped summary rows, and bond coupons must follow the series'' published rates.'
logic_claims:
  - CLAIM-C8022957237352A6
semantic_clauses:
  - 'The demo history must reconcile: every balance assertion it carries must match the balance derived from its transactions, its opening IBKR statement must map with no ingest problem and each monthly one with only the intended skipped summary rows, and bond coupons must follow the series'' published rates.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: adcfcd30231f7fbe01923b6ecc3fc7c308741776872c19f57e0960e6560e3490
semantic_inventory:
  - claim_key: CLAIM-C8022957237352A6
    claim_text: 'The demo history must reconcile: every balance assertion it carries must match the balance derived from its transactions, its opening IBKR statement must map with no ingest problem and each monthly one with only the intended skipped summary rows, and bond coupons must follow the series'' published rates'
    role: normative
    status: modeled
    span:
      start: 0
      end: 303
origin:
  kind: agent
  recorded_at: '2026-10-28T09:31:39.690Z'
id: REQ-demo-consistent-history
type: req
---
The demo history must reconcile: every balance assertion it carries must match the balance derived from its transactions, its opening IBKR statement must map with no ingest problem and each monthly one with only the intended skipped summary rows, and bond coupons must follow the series' published rates.

## Context

A demo full of flagged assertions or unexplained problems would show bugs that aren't there; the simulation tracks cash and units so the statements' ending cash and positions are exactly what the ledger derives (D-18). The summary rows show the ingest-problem feature on purpose. Modeled by Claude Code for the demo-data task: the author asked to see OpenFief on the Hetzner demo with real pipelines on fake data; the author has not reviewed this requirement yet.

## Source

> Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

docs/DESIGN.md D-18
