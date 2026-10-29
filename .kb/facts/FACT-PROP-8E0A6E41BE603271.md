---
title: Plugin Trust Badge Mapping = verified=Verified; signed_by_owner=Yours; signed_by_openfief_ci,signed_by_catalog_publisher,signed_by_unknown_trusted_signer=Unreviewed; refused,not_checked=Blocked
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-87-4
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-87#4
fact_kind: property_value
subject_key: ui.installed_plugins
property_key: plugin_trust_badge_mapping
operator: eq
value_type: string
value_string: verified=Verified; signed_by_owner=Yours; signed_by_openfief_ci,signed_by_catalog_publisher,signed_by_unknown_trusted_signer=Unreviewed; refused,not_checked=Blocked
canonical_key: docs-design-md:ui.installed_plugins:plugin_trust_badge_mapping:eq:verified_verified_signed_by_owner_yours_signed_by_openfief_ci_signed_by_catalog_publisher_signed_by_unknown_trusted_signer_unreviewed_refused_not_checked_blocked
claim_key: CLAIM-95B4B70807F062FF
claim_text: The UI must badge a plugin Verified when it is in the catalog and reviewed, Yours when one of the owner's signers signed it, Unreviewed when another trusted signer signed it and it is not in the catalog, and Blocked when it is refused or not checked yet.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:29:21.532Z'
id: FACT-PROP-8E0A6E41BE603271
type: fact
---
