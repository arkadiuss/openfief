---
title: The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose safety passed by it, the cheapest first.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose safety passed by it, the cheapest first.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 612c9056daacd0e4f16dfa844529531d9ab090f085b0eec1099080b19180f1dd
semantic_inventory:
  - claim_key: CLAIM-9CE76A2C03485355
    claim_text: The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose safety passed by it, the cheapest first
    role: descriptive
    status: missing
    span:
      start: 0
      end: 394
logic_claims:
  - CLAIM-9CE76A2C03485355
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T15:52:41.567Z'
tags:
  - askbench
id: REQ-askbench-routes-and-ranks-by-cost
type: req
---
The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose safety passed by it, the cheapest first.

## Context

The first runs showed one model behaving differently on different hosts, so a run must be able to pin the host. D-207 now picks the demo's model as the cheapest one that passes the Ask box's eval on the hosts that passed it, which is a comparison of what a question costs, not of list prices.

## Source

> The AB3 brief from the design session, 2026-11-14: (2b) host pinning in the bench for the re-run, and its update: cost per question in the report, models ranked by it among those whose safety passed (D-207's cheapest that passes).
