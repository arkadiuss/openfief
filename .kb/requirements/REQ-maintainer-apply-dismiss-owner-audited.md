---
title: Only the owner may apply or dismiss a proposal and each decision is audited in the transaction that makes it, and applying changes the model portfolio's weights only and trades nothing, and is refused when the model portfolio changed since the proposal was made.
status: open
firmness: default
origin:
  kind: agent
  ref: the M3 brief for the model-portfolio maintainer (D-204) on the model router, 2026-11-13
  recorded_at: '2026-11-13T21:07:50.759Z'
semantic_text: Only the owner may apply or dismiss a proposal and each decision is audited in the transaction that makes it, and applying changes the model portfolio's weights only and trades nothing, and is refused when the model portfolio changed since the proposal was made.
semantic_source_field: semantic_text
semantic_source_hash: 20319f5706b500d3290084dc3a801070cbca99b99738a43c9f262973ba0004ea
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-1CBC0345B10084FA
    claim_text: Only the owner may apply or dismiss a proposal and each decision is audited in the transaction that makes it, and applying changes the model portfolio's weights only and trades nothing, and is refused when the model portfolio changed since the proposal was made
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 261
logic_claims:
  - CLAIM-1CBC0345B10084FA
semantic_clauses:
  - Only the owner may apply or dismiss a proposal and each decision is audited in the transaction that makes it, and applying changes the model portfolio's weights only and trades nothing, and is refused when the model portfolio changed since the proposal was made
required_evidence: unverified
id: REQ-maintainer-apply-dismiss-owner-audited
type: req
---
Only the owner may apply or dismiss a proposal and each decision is audited in the transaction that makes it, and applying changes the model portfolio's weights only and trades nothing, and is refused when the model portfolio changed since the proposal was made.

## Context

D-204 has the owner apply or dismiss; D-211 puts decisions in the audit log, which fails closed; the stale check keeps a proposal from overwriting the owner's own edit.

## Source

> Apply (owner; changes the targets, audited) or Dismiss (audited). Applying never trades anything.

Source: the M3 brief for the model-portfolio maintainer (D-204) on the model router, 2026-11-13