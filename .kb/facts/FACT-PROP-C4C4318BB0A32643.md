---
title: Installs Published Release As Guide Says = true
status: active
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.release_smoke
property_key: installs_published_release_as_guide_says
canonical_key: mcp-kibi-model-requirement:tooling.release_smoke:installs_published_release_as_guide_says:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F62483E029A26DC6
claim_text: After every published release, CI must install that release from its GitHub release assets on a fresh kind cluster exactly as the install guide says, through the CLI install of the UI and the first passkey, and fail when any step does not do what the guide says.
origin:
  kind: agent
  recorded_at: '2026-11-06T16:42:35.671Z'
id: FACT-PROP-C4C4318BB0A32643
type: fact
---
