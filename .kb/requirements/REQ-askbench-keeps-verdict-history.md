---
title: The Ask box bench keeps each model@host's runs with their date, verdict, reliability and configuration hash in a history file only when asked, counts consecutive passes and fails over runs of one configuration hash and reports two consecutive passes as eligible for the demo route.
status: open
tags:
  - askbench
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:49.177Z'
semantic_text: The Ask box bench keeps each model@host's runs with their date, verdict, reliability and configuration hash in a history file only when asked, counts consecutive passes and fails over runs of one configuration hash and reports two consecutive passes as eligible for the demo route.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c5242b8e1fb5206328b48a0ee90abc1ca515d0966f5f64740f7ced239785b5a2
semantic_inventory:
  - claim_key: CLAIM-CAE4D83C2CAC082E
    claim_text: The Ask box bench keeps each model@host's runs with their date, verdict, reliability and configuration hash in a history file only when asked, counts consecutive passes and fails over runs of one configuration hash and reports two consecutive passes as eligible for the demo route
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 280
logic_claims:
  - CLAIM-CAE4D83C2CAC082E
id: REQ-askbench-keeps-verdict-history
type: req
---
The Ask box bench keeps each model@host's runs with their date, verdict, reliability and configuration hash in a history file only when asked, counts consecutive passes and fails over runs of one configuration hash and reports two consecutive passes as eligible for the demo route.

## Context

Decided in the AB6 design session (2026-11-14) and the coordinator's D-221 addition: a call that fails on the network says nothing of the model, a host that fails too often should not serve the box whatever its answers, and a model should not join or leave the demo route on a single run.

## Source

> The AB6 brief and the coordinator's D-221 addition, 2026-11-14: Decided in the AB6 design session (2026-11-14) and the coordinator's D-221 addition: a call that fails on the network says nothing of the model, a host that fails too often should not serve the box whatever its answers, and a model should not join or leave the demo route on a single run.
