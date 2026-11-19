---
title: When a model answers in the Ask box, the answer shows one note at its foot with the model's name, how many cards it used or that it used none of the person's data, the line that AI can be wrong, and a link to the page that reports how models did in the eval, names a model the eval has not covered as not evaluated, and says the model's name nowhere else in the answer.
status: open
tags:
  - ask
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: 'Author feedback on demo.openfief.dev from Brave for Android (2026-11-19): a token reported ready but never sent, the check re-running, no progress, too much text; the design session approved the on-demand check and the single attribution note, D-208 and D-222 unchanged'
  recorded_at: '2026-11-19'
semantic_text: When a model answers in the Ask box, the answer shows one note at its foot with the model's name, how many cards it used or that it used none of the person's data, the line that AI can be wrong, and a link to the page that reports how models did in the eval, names a model the eval has not covered as not evaluated, and says the model's name nowhere else in the answer.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 33881b55f3e863c1f7dbee0109f33595d335033394f82d99fe53e3886be11c89
semantic_inventory:
  - claim_key: CLAIM-07955DF152B6858B
    claim_text: When a model answers in the Ask box, the answer shows one note at its foot with the model's name, how many cards it used or that it used none of the person's data, the line that AI can be wrong, and a link to the page that reports how models did in the eval, names a model the eval has not covered as not evaluated, and says the model's name nowhere else in the answer
    role: condition
    status: ontology_gap
    span:
      start: 0
      end: 368
logic_claims:
  - CLAIM-07955DF152B6858B
id: REQ-ui-ask-shows-model-and-disclaimer
type: req
---
When a model answers in the Ask box, the answer shows one note at its foot with the model's name, how many cards it used or that it used none of the person's data, the line that AI can be wrong, and a link to the page that reports how models did in the eval, names a model the eval has not covered as not evaluated, and says the model's name nowhere else in the answer.

## Context

D-222: the person reading an answer should know a model wrote it, that it can be wrong, and where to read how that model did. The commands' own answers come from the record and show none of this.

## Source

> **[D-222]** When a model answers in the Ask box, the answer shows the model's display name from its settings entry, the line "AI can be wrong; check the cards. Not investment advice.", and a link to the docs page that reports how models did in the Ask box's eval (D-217): each model and host with its verdict, failing cases, reliability, cost and run dates, regenerated from the bench's reports. A model the eval hasn't covered is named as not evaluated. Answers to the box's own commands (D-194) show none of this.
