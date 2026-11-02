---
title: Brings Database = true
status: active
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.release_bundle
property_key: brings_database
canonical_key: mcp-kibi-model-requirement:tooling.release_bundle:brings_database:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2C3FBB9F7F3BF878
claim_text: 'Every release bundle must bring its database: CloudNativePG''s release manifest, refused unless its sha256 is the one pinned in the repository, and a minimal openfief-db cluster the owner can edit.'
origin:
  kind: agent
  recorded_at: '2026-11-01T21:51:57.252Z'
id: FACT-PROP-DDE191B64A30136A
type: fact
---
