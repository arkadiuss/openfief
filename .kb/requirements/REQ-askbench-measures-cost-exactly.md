---
title: The Ask box bench reads each call's exact cost (OpenRouter's usage.cost) and serving host from its reply, prices a call from its tokens only when the reply has no cost and says so, lists the hosts that served each model's calls, counts the measured cost against its budget, and ranks the models whose safety passed by measured USD per 1,000 questions with their quality beside it.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box bench reads each call's exact cost (OpenRouter's usage.cost) and serving host from its reply, prices a call from its tokens only when the reply has no cost and says so, lists the hosts that served each model's calls, counts the measured cost against its budget, and ranks the models whose safety passed by measured USD per 1,000 questions with their quality beside it.
tags:
  - askbench
origin:
  kind: agent
  ref: 'AB4 brief from the design session: softer decline wording, injection cases split into safety and answer, reasoning control and exact cost in the bench, 2026-11-14'
  recorded_at: '2026-11-14T16:54:55.083Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1523f56db25ef2d97c39927467f5027002d821dbaf74c54ccf0507491ff0ab6e
semantic_inventory:
  - claim_key: CLAIM-842A1C4CB56C4E79
    claim_text: The Ask box bench reads each call's exact cost (OpenRouter's usage.cost) and serving host from its reply, prices a call from its tokens only when the reply has no cost and says so, lists the hosts that served each model's calls, counts the measured cost against its budget, and ranks the models whose safety passed by measured USD per 1,000 questions with their quality beside it
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 379
logic_claims:
  - CLAIM-842A1C4CB56C4E79
id: REQ-askbench-measures-cost-exactly
type: req
---
The Ask box bench reads each call's exact cost (OpenRouter's usage.cost) and serving host from its reply, prices a call from its tokens only when the reply has no cost and says so, lists the hosts that served each model's calls, counts the measured cost against its budget, and ranks the models whose safety passed by measured USD per 1,000 questions with their quality beside it.

## Context

Token prices ignore what a host really charged and which host served a call. The design session asked for the exact cost and the serving provider per call, a stated fallback to token prices, the hosts per label, and a ranking by measured dollars per 1,000 questions among the models whose safety passed, with quality shown next to it (2026-11-14).

## Source

> AB4 brief, decision (c): send usage: {include: true} on OpenRouter; read usage.cost (USD) and the serving provider from each response; record per call; report cost per question from measured cost (fall back to token x price only when the response lacks it, and say so); report the providers that served each label with call counts. Rank in summary.md by measured $ per 1k questions among models whose SAFETY passed, showing quality next to it.
