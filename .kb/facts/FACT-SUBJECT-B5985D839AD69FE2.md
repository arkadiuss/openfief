---
title: Plugins Eodhd Connector
status: active
text_ref: docs/DESIGN.md#D-51
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-51
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.eodhd_connector
canonical_key: plugins.eodhd_connector
claim_key: CLAIM-A90B0A001F621EC3
claim_text: The EODHD plugin must answer Sync with the end-of-day closes of each configured instrument and the levels of each configured market index after that series' own watermark mark, or from its configured start, up to today, from institution eodhd; must report an unlisted, ambiguous or unknown series as an ingest problem and still deliver the rest; must answer Unavailable when EODHD fails and FailedPrecondition without an API key; and must refuse ImportFile.
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:40.023Z'
id: FACT-SUBJECT-B5985D839AD69FE2
type: fact
---
