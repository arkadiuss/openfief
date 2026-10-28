---
title: Plugins Nbp Fx
status: active
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.nbp_fx
canonical_key: plugins.nbp_fx
claim_key: CLAIM-05D5251DDB9E100A
claim_text: The NBP FX plugin must answer Sync with NBP table A rates to PLN from the day after its watermark, or its configured start on a first sync, up to today, move the watermark to the latest delivered rate's date or keep it when none is new, reach only api.nbp.pl through the egress proxy, and refuse ImportFile.
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:18.339Z'
id: FACT-SUBJECT-C25D641A654E2CC6
type: fact
---
