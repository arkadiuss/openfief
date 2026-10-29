---
title: In Core Pure Functions Behind One Interface = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-70-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-70#1
fact_kind: property_value
subject_key: jurisdiction.pack
property_key: in_core_pure_functions_behind_one_interface
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:jurisdiction.pack:in_core_pure_functions_behind_one_interface:eq:true
claim_key: CLAIM-56C75372137FA3C3
claim_text: In M0 a jurisdiction pack must be a set of pure functions inside the core, behind one interface, so that another jurisdiction is another implementation of it, and must not be a plugin container or a WebAssembly (WASM) module.
origin:
  kind: agent
  recorded_at: '2026-10-29T05:32:56.581Z'
id: FACT-PROP-DF42D044D76C91B0
type: fact
---
