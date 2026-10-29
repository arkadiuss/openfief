---
title: Upgrade Verifies Shows Permission Diff And Awaits Confirmation = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-90-3
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-90#3
fact_kind: property_value
subject_key: plugins.install_cli
property_key: upgrade_verifies_shows_permission_diff_and_awaits_confirmation
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:plugins.install_cli:upgrade_verifies_shows_permission_diff_and_awaits_confirmation:eq:true
claim_key: CLAIM-8494FD797DB3779D
claim_text: openfief plugin upgrade <instance> <version> must verify the new image, show how its requested permissions differ, and apply the upgrade only when the owner confirms.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:35:24.419Z'
id: FACT-PROP-CA08A02C1754E39E
type: fact
---
