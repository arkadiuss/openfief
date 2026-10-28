---
title: For a plugin version, OpenFief must report exactly the requested permissions its grant does not approve, and the version must count as needing approval whenever that report is not empty.
status: open
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
semantic_text: For a plugin version, OpenFief must report exactly the requested permissions its grant does not approve, and the version must count as needing approval whenever that report is not empty.
logic_claims:
  - CLAIM-E4A9CD09256EFEAC
semantic_clauses:
  - For a plugin version, OpenFief must report exactly the requested permissions its grant does not approve, and the version must count as needing approval whenever that report is not empty.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0bcaeb83d84579a9787ad6961b203566b640b4e4b669e06ba28921b5bbf003ce
semantic_inventory:
  - claim_key: CLAIM-E4A9CD09256EFEAC
    claim_text: For a plugin version, OpenFief must report exactly the requested permissions its grant does not approve, and the version must count as needing approval whenever that report is not empty
    role: normative
    status: modeled
    span:
      start: 0
      end: 185
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:47.595Z'
id: REQ-plugins-upgrade-approval-diff
type: req
---
For a plugin version, OpenFief must report exactly the requested permissions its grant does not approve, and the version must count as needing approval whenever that report is not empty.

## Context

D-16 says an upgrade that requests more waits for approval. Reporting the exact unapproved capabilities, egress hosts (with ports), Secret names and data tiers lets the owner approve precisely what changed, and lets the operator keep the old version running meanwhile. A new image digest alone, a smaller request or a moved Secret mount path is not more and needs no approval. Refines REQ-AUTO-8EA5542EFA92FFCD. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> An upgrade that requests more waits for approval.

docs/DESIGN.md D-16
