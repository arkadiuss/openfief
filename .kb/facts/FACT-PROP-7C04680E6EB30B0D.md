---
title: Keeps Api Token Secret = true
status: active
text_ref: AGENTS.md#hard-rules
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:agents-md-hard-rules
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.automation
property_key: keeps_api_token_secret
canonical_key: mcp-kibi-model-requirement:demo.automation:keeps_api_token_secret:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F8C8F64BA3669646
claim_text: The demo automation must read the Hetzner API token from HCLOUD_TOKEN, or from HETZNER_API_TOKEN when HCLOUD_TOKEN is unset, and the Cloudflare token only from CLOUDFLARE_API_TOKEN; it must fail with an explanation when no Hetzner token is set, and must never print either token, pass it as a command argument or write it to disk.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:00:49.127Z'
id: FACT-PROP-7C04680E6EB30B0D
type: fact
---
