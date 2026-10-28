---
title: When no project key is labelled openfief-ssh=true and DEMO_SSH_CIDRS is empty, or DEMO_SSH=off, the demo automation must create the server with a throwaway Ed25519 key labelled app=openfief, env=<env> and openfief-ssh=throwaway whose private half is never written anywhere, reuse that key while it exists, and delete it on down, never deleting a key labelled openfief-ssh=true.
status: open
text_ref: docs/dev/demo.md#ssh
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-ssh
  - lane:strict
semantic_text: When no project key is labelled openfief-ssh=true and DEMO_SSH_CIDRS is empty, or DEMO_SSH=off, the demo automation must create the server with a throwaway Ed25519 key labelled app=openfief, env=<env> and openfief-ssh=throwaway whose private half is never written anywhere, reuse that key while it exists, and delete it on down, never deleting a key labelled openfief-ssh=true.
logic_claims:
  - CLAIM-3CCB37B2F33A8F81
semantic_clauses:
  - When no project key is labelled openfief-ssh=true and DEMO_SSH_CIDRS is empty, or DEMO_SSH=off, the demo automation must create the server with a throwaway Ed25519 key labelled app=openfief, env=<env> and openfief-ssh=throwaway whose private half is never written anywhere, reuse that key while it exists, and delete it on down, never deleting a key labelled openfief-ssh=true.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 830aa5ef1848e24382eb5ab62a908bfdd30c72c4c818ca4462a6ee1aa8236dda
semantic_inventory:
  - claim_key: CLAIM-3CCB37B2F33A8F81
    claim_text: When no project key is labelled openfief-ssh=true and DEMO_SSH_CIDRS is empty, or DEMO_SSH=off, the demo automation must create the server with a throwaway Ed25519 key labelled app=openfief, env=<env> and openfief-ssh=throwaway whose private half is never written anywhere, reuse that key while it exists, and delete it on down, never deleting a key labelled openfief-ssh=true
    role: condition
    status: modeled
    span:
      start: 0
      end: 376
origin:
  kind: agent
  recorded_at: '2026-10-28T14:35:59.101Z'
id: REQ-tooling-demo-ssh-throwaway
type: req
---
When no project key is labelled openfief-ssh=true and DEMO_SSH_CIDRS is empty, or DEMO_SSH=off, the demo automation must create the server with a throwaway Ed25519 key labelled app=openfief, env=<env> and openfief-ssh=throwaway whose private half is never written anywhere, reuse that key while it exists, and delete it on down, never deleting a key labelled openfief-ssh=true.

## Context

Hetzner sets a root password and emails it when a server gets no SSH key. A key nobody holds avoids that password and leaves no way in, so the demo is managed only through git and Flux. Modeled by Claude Code for the demo machine automation. The author asked, through the orchestrating agent, for the demo on the real domain openfief.dev (a Cloudflare zone on the Free plan), fully automated, and for an explicit SSH-off mode because the Hetzner project has no SSH key and Hetzner would otherwise email a root password. The docs site shares the zone (docs.openfief.dev), so the demo must touch only its own records. The author has not reviewed this requirement yet.

## Source

> Add an explicit, documented DEMO_SSH=off mode ... That way nobody has a root password and nobody can SSH; the box is managed only via GitOps.

Orchestrator brief for impl/demo-domain; docs/dev/demo.md#ssh
