---
title: Seeds Only In Opening Mode = true
status: active
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ingest_path
property_key: seeds_only_in_opening_mode
canonical_key: mcp-kibi-model-requirement:ingest.ingest_path:seeds_only_in_opening_mode:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-49C4C404000888F9
claim_text: An opening import batch must seed through the ledger's seeding use case, which refuses an already seeded account, and a transactions batch that holds opening lots or opening cash must be refused before anything is imported.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:39:32.543Z'
id: FACT-PROP-07F196F017D74FAE
type: fact
---
