---
title: 'The owner''s balance check must show, for each balance assertion, the reported balance, the balance derived from the transaction log, reported minus derived, and whether it passes: a cash assertion within 0.01 of its currency either way, a holding assertion only with equal units.'
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: 'The owner''s balance check must show, for each balance assertion, the reported balance, the balance derived from the transaction log, reported minus derived, and whether it passes: a cash assertion within 0.01 of its currency either way, a holding assertion only with equal units.'
logic_claims:
  - CLAIM-206B121C6E083E44
semantic_clauses:
  - 'The owner''s balance check must show, for each balance assertion, the reported balance, the balance derived from the transaction log, reported minus derived, and whether it passes: a cash assertion within 0.01 of its currency either way, a holding assertion only with equal units.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ceb55c2c4db8dfb0b11416b149f80d6ca36f15ffea31442dfa4bc6fad428db6a
semantic_inventory:
  - claim_key: CLAIM-206B121C6E083E44
    claim_text: 'The owner''s balance check must show, for each balance assertion, the reported balance, the balance derived from the transaction log, reported minus derived, and whether it passes: a cash assertion within 0.01 of its currency either way, a holding assertion only with equal units'
    role: normative
    status: modeled
    span:
      start: 0
      end: 278
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:20.212Z'
id: REQ-views-balance-check-tolerance
type: req
---
The owner's balance check must show, for each balance assertion, the reported balance, the balance derived from the transaction log, reported minus derived, and whether it passes: a cash assertion within 0.01 of its currency either way, a holding assertion only with equal units.

## Context

D-200 criterion 1 is the owner's own check on a private install with real accounts, never in CI (GR-12). The design session asked, through the orchestrator on 2026-11-06, for an owner-only command, `openfief check balances`, so the author can check it: every balance assertion with reported and derived values, the difference and pass or fail at 0.01 of the account currency, open XTB CFDs flagged as unsupported (D-24), values printed only to the owner's terminal, the core logging only that the check ran with counts, a non-zero exit status on a failure, a table and --json. Reading 0.01 of the account currency as 0.01 of each cash assertion's own currency, and requiring equal units for holdings, whose units carry no currency, is this context's reading.

## Source

> Every broker-reported position and cash balance is stored as a balance assertion and matches the derived value to 0.01 of the account currency. Open XTB CFDs appear only as flagged, unsupported items.

docs/DESIGN.md#D-200
