---
title: Resolves Catalog Or Pinned Image = true
status: active
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.install_cli
property_key: resolves_catalog_or_pinned_image
canonical_key: mcp-kibi-model-requirement:plugins.install_cli:resolves_catalog_or_pinned_image:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F0E6462C065C0EAA
claim_text: openfief plugin install must take either the name of exactly one catalog entry or a digest-pinned image reference, must find the catalog entry that pins an image's digest, must refuse a tag, an unknown name or a name several entries share, and must name the instance after the catalog entry or the image repository unless the owner names it.
origin:
  kind: agent
  recorded_at: '2026-10-28T22:24:01.543Z'
id: FACT-PROP-FF2109D7BECBB423
type: fact
---
