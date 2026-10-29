---
title: Max File Size Mib <= 64
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-109-6
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-109#6
fact_kind: property_value
subject_key: ingest.statement_upload
property_key: max_file_size_mib
operator: lte
value_type: int
value_int: 64
canonical_key: docs-design-md:ingest.statement_upload:max_file_size_mib:lte:64
claim_key: CLAIM-E0A068FCB7D19A88
claim_text: The core must accept statement files up to 64 MiB.
origin:
  kind: agent
  recorded_at: '2026-10-29T10:50:30.441Z'
id: FACT-PROP-6B6FF86E0CD2B486
type: fact
---
