---
title: Signature Trust Levels = not_checked, refused, signed_by_catalog_publisher, signed_by_openfief_ci, signed_by_owner, signed_by_unknown_trusted_signer, verified
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-87-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-87#1
fact_kind: property_value
subject_key: operator.plugin_status
property_key: signature_trust_levels
operator: eq
value_type: set
value_set:
  - not_checked
  - refused
  - signed_by_catalog_publisher
  - signed_by_openfief_ci
  - signed_by_owner
  - signed_by_unknown_trusted_signer
  - verified
canonical_key: docs-design-md:operator.plugin_status:signature_trust_levels:eq:not_checked_refused_signed_by_catalog_publisher_signed_by_openfief_ci_signed_by_owner_signed_by_unknown_trusted_signer_verified
claim_key: CLAIM-7DAAFBDE0635CA7C
claim_text: 'The operator must record each plugin''s signature trust in its status in full: verified, signed by OpenFief''s CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, or not checked.'
origin:
  kind: agent
  recorded_at: '2026-10-29T08:27:16.675Z'
id: FACT-PROP-DD74EC2102737DA1
type: fact
---
