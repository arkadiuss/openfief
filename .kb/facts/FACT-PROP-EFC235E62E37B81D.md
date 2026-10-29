---
title: Unprefixed Session Cookie Hosts within localhost
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-85-3
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-85#3
fact_kind: property_value
subject_key: views.built_in_ui
property_key: unprefixed_session_cookie_hosts
operator: subset_of
value_type: set
value_set:
  - localhost
canonical_key: docs-design-md:views.built_in_ui:unprefixed_session_cookie_hosts:subset_of:localhost
claim_key: CLAIM-FA40DAA3AC4F2570
claim_text: A session cookie without the __Host- prefix must be used only on localhost, and only if Safari refuses the prefixed cookie there.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:24:07.183Z'
id: FACT-PROP-EFC235E62E37B81D
type: fact
---
