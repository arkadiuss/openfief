---
title: The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose verdict (D-220) and host reliability (D-207) both pass by it, the cheapest first.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose verdict (D-220) and host reliability (D-207) both pass by it, the cheapest first.
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:58.941Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e2b88184d11fdc359e8e1689b6901dff869766f9b02f85a9a27196b8e0b0d41c
semantic_inventory:
  - claim_key: CLAIM-D33914ED372D105F
    claim_text: The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose verdict (D-220) and host reliability (D-207) both pass by it, the cheapest first
    role: descriptive
    status: missing
    span:
      start: 0
      end: 435
logic_claims:
  - CLAIM-D33914ED372D105F
id: REQ-askbench-routes-and-ranks-by-cost-v2
type: req
tags:
  - askbench
---
The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose verdict (D-220) and host reliability (D-207) both pass by it, the cheapest first.

## Context

The first runs showed one model behaving differently on different hosts, so a run must be able to pin the host. D-207 picks the demo's model as the cheapest one that passes the Ask box's eval on the hosts that passed it; D-220 says what passing is, so the bench ranks the models whose verdict is pass and no longer those whose safety passed alone, which a model that refuses everything also does.

## Why this changed

The first real runs showed a host that failed transport calls often enough that a model on it could not serve the box however well it answered when it answered. The design session decided that a host over five percent transport failures fails reliability for the run, so the ranking lists a model only when its verdict and its host's reliability both pass.

## Why this changed

D-220 now says what passing the Ask box's eval means, and the bench gives each model@host a verdict. The ranking is therefore of the models whose verdict is pass, no longer of those whose safety alone passed, which a model that refuses everything also does.

## Source

> The AB6 brief from the design session (2026-11-14): report the transport-failure rate per model@host per run; a host over 5% fails reliability whatever its verdict; the D-207 ranking lists only models@hosts whose verdict and reliability pass.
