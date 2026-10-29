---
title: Plugins Ghostfolio Plugin
status: active
text_ref: docs/DESIGN.md#D-3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-3
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.ghostfolio_plugin
canonical_key: plugins.ghostfolio_plugin
claim_key: CLAIM-A4F7C0D5B8862902
claim_text: The Ghostfolio plugin's manifest must ask only for the view capability and the ghostfolio-credentials Secret holding the Ghostfolio security token; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach Ghostfolio only through the egress proxy and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.
origin:
  kind: agent
  recorded_at: '2026-10-29T01:36:56.808Z'
id: FACT-SUBJECT-D5E66895C31CDA5F
type: fact
---
