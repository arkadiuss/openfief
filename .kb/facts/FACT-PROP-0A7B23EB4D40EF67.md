---
title: Bond Asset Identity = MANUAL asset, symbol UUID from plugin instance and series code, name is the series code
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-6
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-64#6
fact_kind: property_value
subject_key: views.ghostfolio_plugin
property_key: bond_asset_identity
operator: eq
value_type: string
value_string: MANUAL asset, symbol UUID from plugin instance and series code, name is the series code
canonical_key: docs-design-md:views.ghostfolio_plugin:bond_asset_identity:eq:manual_asset_symbol_uuid_from_plugin_instance_and_series_code_name_is_the_series_code
claim_key: CLAIM-26D63061DACA79C0
claim_text: The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code, with the series code as its name.
origin:
  kind: agent
  recorded_at: '2026-10-29T10:23:20.971Z'
id: FACT-PROP-0A7B23EB4D40EF67
type: fact
---
