---
title: The Ask box's system prompt tells the model to refuse only a change, advice on what to buy, sell or hold, and a question that is not about the portfolio, to greet back when greeted and to answer every other question about the portfolio, why-questions included, from the cards, and to refuse in one friendly sentence, with the canary and the split of public and non-public sentences intact.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box's system prompt tells the model to refuse only a change, advice on what to buy, sell or hold, and a question that is not about the portfolio, to greet back when greeted and to answer every other question about the portfolio, why-questions included, from the cards, and to refuse in one friendly sentence, with the canary and the split of public and non-public sentences intact.
tags:
  - ask
origin:
  kind: agent
  ref: 'AB4 brief from the design session: softer decline wording, injection cases split into safety and answer, reasoning control and exact cost in the bench, 2026-11-14'
  recorded_at: '2026-11-14T16:54:58.839Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e86f62c2d8d049724e593789d89a5e72b2f59a51e33d5b5eba7310f53d651b2c
semantic_inventory:
  - claim_key: CLAIM-7151EB2DF1C0454A
    claim_text: The Ask box's system prompt tells the model to refuse only a change, advice on what to buy, sell or hold, and a question that is not about the portfolio, to greet back when greeted and to answer every other question about the portfolio, why-questions included, from the cards, and to refuse in one friendly sentence, with the canary and the split of public and non-public sentences intact
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 388
logic_claims:
  - CLAIM-7151EB2DF1C0454A
id: REQ-ask-prompt-declines-only-three-things
type: req
---
The Ask box's system prompt tells the model to refuse only a change, advice on what to buy, sell or hold, and a question that is not about the portfolio, to greet back when greeted and to answer every other question about the portfolio, why-questions included, from the cards, and to refuse in one friendly sentence, with the canary and the split of public and non-public sentences intact.

## Context

The first live run showed a model declining questions it should have answered; the design session decided the decline rule names exactly three classes, softens the wording and answers the rest (2026-11-14).

## Source

> AB4 brief, decision (a): PROMPT: refuse ONLY changes (writes), buy/sell advice, and questions outside the portfolio. Greet; answer portfolio questions normally, including "why" questions answered from the cards. Keep the canary and the public/non-public split intact. Add a prompt test that the decline rule names exactly those three classes.
