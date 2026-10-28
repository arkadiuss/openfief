---
title: Same Transaction As Input = true
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.enqueue
property_key: same_transaction_as_input
canonical_key: mcp-kibi-model-requirement:valuation.enqueue:same_transaction_as_input:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-97718BFE48BD8939
claim_text: Storing a price, FX rate, index reading, bond series terms or tax wrapper, and every ledger import, must enqueue its revaluations in the same database transaction, so that a rollback undoes both.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:57:14.573Z'
id: FACT-PROP-FEC85F856CD3B3AC
type: fact
---
