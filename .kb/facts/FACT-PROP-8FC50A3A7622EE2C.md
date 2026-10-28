---
title: New Hostname Via Session Link = true
status: active
text_ref: design-decisions:D-60#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-60-2
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: identity.passkey_enrollment
property_key: new_hostname_via_session_link
canonical_key: docs-design-md:identity.passkey_enrollment:new_hostname_via_session_link:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-207A16FB308FD2F2
claim_text: A logged-in session must be able to issue a one-time link for creating a passkey on a new hostname, and passkeys must be kept per hostname.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:46.481Z'
id: FACT-PROP-8FC50A3A7622EE2C
type: fact
---
