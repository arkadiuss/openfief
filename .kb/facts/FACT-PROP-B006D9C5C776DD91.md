---
title: Own Schema Hash Only = true
status: active
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: identity.script_token_store
property_key: own_schema_hash_only
canonical_key: mcp-kibi-model-requirement:identity.script_token_store:own_schema_hash_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-95DD9512447C8C06
claim_text: The identity context must keep script tokens in its own PostgreSQL schema, migrated by its own runner and lock, with a row per token holding its ID, name, scope, times and the 32-byte SHA-256 hash of its key, and must refuse rows with a malformed ID, name or hash or an expiry not after creation.
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:35.640Z'
id: FACT-PROP-B006D9C5C776DD91
type: fact
---
