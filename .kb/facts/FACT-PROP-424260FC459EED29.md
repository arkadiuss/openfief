---
title: Report Shows E2e Receipts = true
status: active
text_ref: docs/dev/kibi.md#evidence-as-kibi-counts-it
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-kibi-md-evidence-as-kibi-counts-it
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.docs_site
property_key: report_shows_e2e_receipts
canonical_key: mcp-kibi-model-requirement:tooling.docs_site:report_shows_e2e_receipts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-41C13B43A8C5CDA1
claim_text: The docs site published from main must be built only after both end-to-end runs passed, with each run's kibi.proof-run.v1 artifact ingested by kibi into the same commit's knowledge base, oldest run first, and must fail its build when kibi refuses an artifact.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:45:44.178Z'
id: FACT-PROP-424260FC459EED29
type: fact
---
