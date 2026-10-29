---
title: Bounded And Untrusted Metadata = true
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.statement_upload
property_key: bounded_and_untrusted_metadata
canonical_key: mcp-kibi-model-requirement:ingest.statement_upload:bounded_and_untrusted_metadata:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C2906C4F54F95FE2
claim_text: An uploaded statement must be refused before a connector reads it when the caller is not the owner, no mode is given or the file is empty, and stopped when it passes 64 MiB; the connector gets only the cleaned last element of its name and a sniffed media type.
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:25.478Z'
id: FACT-PROP-4CDF356EC8A0844C
type: fact
---
