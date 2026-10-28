---
title: When the current version of a plugin requests permissions its grant does not approve, the previously deployed version must keep running under the current grant, and nothing must run when no version was deployed before.
status: open
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
semantic_text: When the current version of a plugin requests permissions its grant does not approve, the previously deployed version must keep running under the current grant, and nothing must run when no version was deployed before.
logic_claims:
  - CLAIM-E059A7B53B3729C9
semantic_clauses:
  - When the current version of a plugin requests permissions its grant does not approve, the previously deployed version must keep running under the current grant, and nothing must run when no version was deployed before.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7fd7c3daf87b4a2c3a8725f64530577155cdeb9d098f4e6fa259c6fd4b0b39bc
semantic_inventory:
  - claim_key: CLAIM-E059A7B53B3729C9
    claim_text: When the current version of a plugin requests permissions its grant does not approve, the previously deployed version must keep running under the current grant, and nothing must run when no version was deployed before
    role: condition
    status: modeled
    span:
      start: 0
      end: 217
origin:
  kind: agent
  recorded_at: '2026-10-28T06:36:29.709Z'
id: REQ-plugins-rollout-upgrade-waits
type: req
---
When the current version of a plugin requests permissions its grant does not approve, the previously deployed version must keep running under the current grant, and nothing must run when no version was deployed before.

## Context

D-16 says an upgrade that requests more waits for approval and the operator enforces only what is in both the Plugin and the PluginGrant. Our reading: waiting means the version that already runs keeps running, limited to the current grant, so a narrowed grant takes effect at once; a first install that requests more than its grant runs nothing, and a current version refused by a check outside the domain (image signature, config schema) is treated the same way. Refines REQ-AUTO-8EA5542EFA92FFCD. Modeled by Claude Code for the M0 operator task; the author has not reviewed it yet.

## Source

> An upgrade that requests more waits for approval.

docs/DESIGN.md D-16
