---
title: Reports Removed Duplicates = true
status: active
text_ref: docs/dev/kibi.md#setup-notes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-kibi-md-setup-notes
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.kibi_gate
property_key: reports_removed_duplicates
canonical_key: mcp-kibi-model-requirement:tooling.kibi_gate:reports_removed_duplicates:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1B078F0E6E0509D3
claim_text: When kibi's engine attaches a published store holding duplicate triples, it must keep one copy of each, and the next kibi command must report how many duplicate triples and entities stored twice it removed, as a workflow warning in CI.
origin:
  kind: agent
  recorded_at: '2026-11-07T19:28:03.651Z'
id: FACT-PROP-6DB41F9D1C53C474
type: fact
---
