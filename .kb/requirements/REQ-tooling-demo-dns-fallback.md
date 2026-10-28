---
title: Without CLOUDFLARE_API_TOKEN, the demo automation must make no DNS API call, must say that DNS is not managed and print the A record to create, without failing; with DEMO_DOMAIN=none it must skip DNS entirely and let the machine use its <ipv4-with-dashes>.sslip.io name.
status: open
text_ref: docs/dev/demo.md#dns
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-dns
  - lane:strict
semantic_text: Without CLOUDFLARE_API_TOKEN, the demo automation must make no DNS API call, must say that DNS is not managed and print the A record to create, without failing; with DEMO_DOMAIN=none it must skip DNS entirely and let the machine use its <ipv4-with-dashes>.sslip.io name.
logic_claims:
  - CLAIM-F3F96B05788D4B93
semantic_clauses:
  - Without CLOUDFLARE_API_TOKEN, the demo automation must make no DNS API call, must say that DNS is not managed and print the A record to create, without failing; with DEMO_DOMAIN=none it must skip DNS entirely and let the machine use its <ipv4-with-dashes>.sslip.io name.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6734bd1c5610f07ad727d968e0356c93bcc5467e62487c9fcbd42a679ba62045
semantic_inventory:
  - claim_key: CLAIM-F3F96B05788D4B93
    claim_text: Without CLOUDFLARE_API_TOKEN, the demo automation must make no DNS API call, must say that DNS is not managed and print the A record to create, without failing; with DEMO_DOMAIN=none it must skip DNS entirely and let the machine use its <ipv4-with-dashes>.sslip.io name
    role: normative
    status: modeled
    span:
      start: 0
      end: 269
origin:
  kind: agent
  recorded_at: '2026-10-28T14:35:53.280Z'
id: REQ-tooling-demo-dns-fallback
type: req
---
Without CLOUDFLARE_API_TOKEN, the demo automation must make no DNS API call, must say that DNS is not managed and print the A record to create, without failing; with DEMO_DOMAIN=none it must skip DNS entirely and let the machine use its <ipv4-with-dashes>.sslip.io name.

## Context

A session or fork without the Cloudflare token can still create a demo and set the record by hand, as the design agent's first version did; the sslip.io name needs no DNS record at all. Modeled by Claude Code for the demo machine automation. The author asked, through the orchestrating agent, for the demo on the real domain openfief.dev (a Cloudflare zone on the Free plan), fully automated, and for an explicit SSH-off mode because the Hetzner project has no SSH key and Hetzner would otherwise email a root password. The docs site shares the zone (docs.openfief.dev), so the demo must touch only its own records. The author has not reviewed this requirement yet.

## Source

> The author owns the Cloudflare zone openfief.dev and wants the demo on a real domain, fully automated ("remember about automation").

Orchestrator brief for impl/demo-domain; docs/dev/demo.md#dns
