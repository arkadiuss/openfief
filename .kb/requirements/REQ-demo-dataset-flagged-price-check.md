---
title: The demo dataset must carry synthetic IBKR mark prices as cross-check prices at each monthly statement's end, exactly one of them beyond the 1% threshold, and the demo loader must record them, so the demo's price checks page shows a flagged check.
status: open
text_ref: docs/DESIGN.md#D-110
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-110
  - lane:strict
semantic_text: The demo dataset must carry synthetic IBKR mark prices as cross-check prices at each monthly statement's end, exactly one of them beyond the 1% threshold, and the demo loader must record them, so the demo's price checks page shows a flagged check.
logic_claims:
  - CLAIM-D234E5E8B0D52BB4
semantic_clauses:
  - The demo dataset must carry synthetic IBKR mark prices as cross-check prices at each monthly statement's end, exactly one of them beyond the 1% threshold, and the demo loader must record them, so the demo's price checks page shows a flagged check.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 38545fa91ac1046e97bf64f93b33d51b95b974b4a5d1d44fbaab95c29887fe4a
semantic_inventory:
  - claim_key: CLAIM-D234E5E8B0D52BB4
    claim_text: The demo dataset must carry synthetic IBKR mark prices as cross-check prices at each monthly statement's end, exactly one of them beyond the 1% threshold, and the demo loader must record them, so the demo's price checks page shows a flagged check
    role: normative
    status: modeled
    span:
      start: 0
      end: 246
origin:
  kind: agent
  recorded_at: '2026-10-30T05:18:23.949Z'
id: REQ-demo-dataset-flagged-price-check
type: req
---
The demo dataset must carry synthetic IBKR mark prices as cross-check prices at each monthly statement's end, exactly one of them beyond the 1% threshold, and the demo loader must record them, so the demo's price checks page shows a flagged check.

## Context

docs/dev/demo.md says the dataset carries one cross-check price beyond the threshold so the demo shows a flagged check; the design session found the demo's page showed none.

## Source

> A cross-check price is flagged when it differs from the stored price by more than 1% of it, in either direction.

docs/DESIGN.md#D-110
