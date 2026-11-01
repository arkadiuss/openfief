---
title: Matches Channel Format = true
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
subject_key: release.version
property_key: matches_channel_format
canonical_key: mcp-kibi-model-requirement:release.version:matches_channel_format:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4CFFF825D4F8EE6D
claim_text: The release workflow must refuse a version that does not match its channel's format, YYYY.M.N for stable, YYYY.M.N-beta.K for beta and YYYY.M.N-alpha.K for alpha, with no v prefix, no leading zeros and K from 1, or whose tag already exists.
origin:
  kind: agent
  recorded_at: '2026-11-01T14:58:32.917Z'
id: FACT-PROP-3231FAEBB5DFEC86
type: fact
---
