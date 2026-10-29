---
title: Syncs On A Timer With Its Grant = true
status: active
text_ref: docs/DESIGN.md#D-3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-3
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.ghostfolio_plugin
property_key: syncs_on_a_timer_with_its_grant
canonical_key: mcp-kibi-model-requirement:plugins.ghostfolio_plugin:syncs_on_a_timer_with_its_grant:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A4F7C0D5B8862902
claim_text: The Ghostfolio plugin's manifest must ask only for the view capability and the ghostfolio-credentials Secret holding the Ghostfolio security token; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach Ghostfolio only through the egress proxy and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.
origin:
  kind: agent
  recorded_at: '2026-10-29T01:36:58.764Z'
id: FACT-PROP-C36484CF5E5B360E
type: fact
---
