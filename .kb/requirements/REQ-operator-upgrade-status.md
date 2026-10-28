---
title: When a Plugin's spec requests permissions its PluginGrant does not approve, the operator must keep running the previously deployed version and report exactly the unapproved permissions in status.pendingApproval and an ApprovalRequired condition.
status: open
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
semantic_text: When a Plugin's spec requests permissions its PluginGrant does not approve, the operator must keep running the previously deployed version and report exactly the unapproved permissions in status.pendingApproval and an ApprovalRequired condition.
logic_claims:
  - CLAIM-5D427B563C4C0A91
semantic_clauses:
  - When a Plugin's spec requests permissions its PluginGrant does not approve, the operator must keep running the previously deployed version and report exactly the unapproved permissions in status.pendingApproval and an ApprovalRequired condition.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d785fa09ec7459ee497ce8f137534764e741ee3e15a67a4ab835e5d6137669b3
semantic_inventory:
  - claim_key: CLAIM-5D427B563C4C0A91
    claim_text: When a Plugin's spec requests permissions its PluginGrant does not approve, the operator must keep running the previously deployed version and report exactly the unapproved permissions in status.pendingApproval and an ApprovalRequired condition
    role: condition
    status: modeled
    span:
      start: 0
      end: 244
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:15.469Z'
id: REQ-operator-upgrade-status
type: req
---
When a Plugin's spec requests permissions its PluginGrant does not approve, the operator must keep running the previously deployed version and report exactly the unapproved permissions in status.pendingApproval and an ApprovalRequired condition.

## Context

D-16 has an upgrade that requests more wait for approval. The owner needs to see what to approve, so the difference is in the status and condition message; the previously deployed version is recorded in status.deployed, because it is what keeps running. Refines REQ-AUTO-8EA5542EFA92FFCD. Modeled by Claude Code for the M0 operator task; the author has not reviewed it yet.

## Source

> An upgrade that requests more waits for approval.

docs/DESIGN.md D-16
