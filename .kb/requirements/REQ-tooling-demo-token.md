---
title: The demo automation must read the Hetzner API token from HCLOUD_TOKEN, or from HETZNER_API_TOKEN when HCLOUD_TOKEN is unset, and the Cloudflare token only from CLOUDFLARE_API_TOKEN; it must fail with an explanation when no Hetzner token is set, and must never print either token, pass it as a command argument or write it to disk.
status: open
text_ref: AGENTS.md#hard-rules
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:agents-md-hard-rules
  - lane:strict
semantic_text: The demo automation must read the Hetzner API token from HCLOUD_TOKEN, or from HETZNER_API_TOKEN when HCLOUD_TOKEN is unset, and the Cloudflare token only from CLOUDFLARE_API_TOKEN; it must fail with an explanation when no Hetzner token is set, and must never print either token, pass it as a command argument or write it to disk.
logic_claims:
  - CLAIM-F8C8F64BA3669646
semantic_clauses:
  - The demo automation must read the Hetzner API token from HCLOUD_TOKEN, or from HETZNER_API_TOKEN when HCLOUD_TOKEN is unset, and the Cloudflare token only from CLOUDFLARE_API_TOKEN; it must fail with an explanation when no Hetzner token is set, and must never print either token, pass it as a command argument or write it to disk.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c8efacff85ea6e139d869db1daf0d689939412d7fdeea4284cb8a1251e61479b
semantic_inventory:
  - claim_key: CLAIM-F8C8F64BA3669646
    claim_text: The demo automation must read the Hetzner API token from HCLOUD_TOKEN, or from HETZNER_API_TOKEN when HCLOUD_TOKEN is unset, and the Cloudflare token only from CLOUDFLARE_API_TOKEN; it must fail with an explanation when no Hetzner token is set, and must never print either token, pass it as a command argument or write it to disk
    role: normative
    status: modeled
    span:
      start: 0
      end: 329
origin:
  kind: agent
  recorded_at: '2026-10-28T09:00:53.466Z'
id: REQ-tooling-demo-token
type: req
---
The demo automation must read the Hetzner API token from HCLOUD_TOKEN, or from HETZNER_API_TOKEN when HCLOUD_TOKEN is unset, and the Cloudflare token only from CLOUDFLARE_API_TOKEN; it must fail with an explanation when no Hetzner token is set, and must never print either token, pass it as a command argument or write it to disk.

## Context

Both tokens can create and delete resources in the owner's accounts, and the repository and its Actions logs are public. The hard rules forbid credentials in code, logs and commits; the scripts also ignore stored hcloud CLI contexts, and pass the Cloudflare token to curl on standard input. Modeled by Claude Code for the demo machine automation. The author asked, through the orchestrating agent, for the demo on the real domain openfief.dev (a Cloudflare zone on the Free plan), fully automated, and for an explicit SSH-off mode because the Hetzner project has no SSH key and Hetzner would otherwise email a root password. The docs site shares the zone (docs.openfief.dev), so the demo must touch only its own records. The author has not reviewed this requirement yet.

## Source

> No real financial data, account numbers or credentials anywhere: code, tests, recordings, logs, commits.

AGENTS.md#hard-rules
