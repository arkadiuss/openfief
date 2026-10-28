---
title: task dns:docs must make docs.<DEMO_DOMAIN> a DNS-only CNAME to the GitHub Pages host with the comment managed-by=openfief-docs, creating or updating only a record with that comment, refusing when any other record occupies the name, and task dns:docs:plan must show the change without making it; the Cloudflare token is handled as for the demo.
status: open
text_ref: docs/dev/demo.md#dns
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-dns
  - lane:strict
semantic_text: task dns:docs must make docs.<DEMO_DOMAIN> a DNS-only CNAME to the GitHub Pages host with the comment managed-by=openfief-docs, creating or updating only a record with that comment, refusing when any other record occupies the name, and task dns:docs:plan must show the change without making it; the Cloudflare token is handled as for the demo.
logic_claims:
  - CLAIM-044F396F49BFC390
semantic_clauses:
  - task dns:docs must make docs.<DEMO_DOMAIN> a DNS-only CNAME to the GitHub Pages host with the comment managed-by=openfief-docs, creating or updating only a record with that comment, refusing when any other record occupies the name, and task dns:docs:plan must show the change without making it; the Cloudflare token is handled as for the demo.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8bd0b4ea7b01226f2a944f70b83ec95d9889c421adbdfbd942d5976c455427be
semantic_inventory:
  - claim_key: CLAIM-044F396F49BFC390
    claim_text: task dns:docs must make docs.<DEMO_DOMAIN> a DNS-only CNAME to the GitHub Pages host with the comment managed-by=openfief-docs, creating or updating only a record with that comment, refusing when any other record occupies the name, and task dns:docs:plan must show the change without making it; the Cloudflare token is handled as for the demo
    role: normative
    status: modeled
    span:
      start: 0
      end: 342
origin:
  kind: agent
  recorded_at: '2026-10-28T14:36:20.203Z'
id: REQ-tooling-demo-dns-docs
type: req
---
task dns:docs must make docs.<DEMO_DOMAIN> a DNS-only CNAME to the GitHub Pages host with the comment managed-by=openfief-docs, creating or updating only a record with that comment, refusing when any other record occupies the name, and task dns:docs:plan must show the change without making it; the Cloudflare token is handled as for the demo.

## Context

The docs site moves to docs.openfief.dev on GitHub Pages in the same Cloudflare zone as the demo. The orchestrating agent asked for a small automated, idempotent target for it that reuses the demo's Cloudflare helpers; separate comments keep the demo and docs records from ever touching each other. Modeled by Claude Code for the demo machine automation. The author asked, through the orchestrating agent, for the demo on the real domain openfief.dev (a Cloudflare zone on the Free plan), fully automated, and for an explicit SSH-off mode because the Hetzner project has no SSH key and Hetzner would otherwise email a root password. The docs site shares the zone (docs.openfief.dev), so the demo must touch only its own records. The author has not reviewed this requirement yet.

## Source

> The author owns the Cloudflare zone openfief.dev and wants the demo on a real domain, fully automated ("remember about automation").

Orchestrator brief for impl/demo-domain; docs/dev/demo.md#dns
