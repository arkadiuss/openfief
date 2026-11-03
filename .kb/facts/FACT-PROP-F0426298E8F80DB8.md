---
title: No Values Logged = true
status: active
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.logging
property_key: no_values_logged
canonical_key: mcp-kibi-model-requirement:ingest.logging:no_values_logged:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0A05AB1AE74B1EA8
claim_text: The ingest context must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry, and the ledger's refusals of a transaction must not quote its amount or quantity.
origin:
  kind: agent
  recorded_at: '2026-11-03T01:05:23.888Z'
id: FACT-PROP-F0426298E8F80DB8
type: fact
---
