---
title: Fails On Covered Change Without Page Or Trailer = true
status: active
text_ref: docs/dev/docs.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-docs-md
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: docs.stale_gate
property_key: fails_on_covered_change_without_page_or_trailer
canonical_key: docs-dev-docs-md:docs.stale_gate:fails_on_covered_change_without_page_or_trailer:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EC277DB6E1B2445F
claim_text: A change to a file that a docs page covers must also change a page covering that file in the same push, or carry a Docs-Unchanged trailer; tests and test data never count.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:32:48.908Z'
id: FACT-PROP-3CDD13D16A2E3D9D
type: fact
---
