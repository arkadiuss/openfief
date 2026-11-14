---
title: The Ask box bench reads the demo route's members from its config without editing it, asks them when no model is named, exits with a failure when a member's run fails its verdict or its reliability and reports the members with two consecutive failing runs for removal in a reviewed change.
status: open
tags:
  - askbench
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:51.107Z'
semantic_text: The Ask box bench reads the demo route's members from its config without editing it, asks them when no model is named, exits with a failure when a member's run fails its verdict or its reliability and reports the members with two consecutive failing runs for removal in a reviewed change.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c877904b0c051b513e72ec5dd834ac759f1b283bd73fbe6cf603c8d187a20f38
semantic_inventory:
  - claim_key: CLAIM-556405216442E05E
    claim_text: The Ask box bench reads the demo route's members from its config without editing it, asks them when no model is named, exits with a failure when a member's run fails its verdict or its reliability and reports the members with two consecutive failing runs for removal in a reviewed change
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 287
logic_claims:
  - CLAIM-556405216442E05E
id: REQ-askbench-gates-the-demo-route
type: req
---
The Ask box bench reads the demo route's members from its config without editing it, asks them when no model is named, exits with a failure when a member's run fails its verdict or its reliability and reports the members with two consecutive failing runs for removal in a reviewed change.

## Context

Decided in the AB6 design session (2026-11-14) and the coordinator's D-221 addition: a call that fails on the network says nothing of the model, a host that fails too often should not serve the box whatever its answers, and a model should not join or leave the demo route on a single run.

## Source

> The AB6 brief and the coordinator's D-221 addition, 2026-11-14: Decided in the AB6 design session (2026-11-14) and the coordinator's D-221 addition: a call that fails on the network says nothing of the model, a host that fails too often should not serve the box whatever its answers, and a model should not join or leave the demo route on a single run.
