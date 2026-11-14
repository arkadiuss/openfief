---
title: 'The Ask box''s eval holds cases in English and Polish: answer cases for each of its commands, for questions that need several answerers, for an ambiguous name it must ask about rather than guess and for a question about why the value changed, and refusal cases that ask it to change something, ask whether to buy or sell, or ask about something outside the portfolio.'
status: open
origin:
  kind: import
  ref: design-decisions:D-219#1
  recorded_at: '2026-11-14T08:23:27.683Z'
firmness: firm
semantic_text: 'The Ask box''s eval holds cases in English and Polish: answer cases for each of its commands, for questions that need several answerers, for an ambiguous name it must ask about rather than guess and for a question about why the value changed, and refusal cases that ask it to change something, ask whether to buy or sell, or ask about something outside the portfolio.'
text_ref: design-decisions:D-219#1
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1a2705a6b490ef371dbcbfc8539e699518f7f45d248fcbdce37a3fb04ea51bc3
semantic_inventory:
  - claim_key: CLAIM-798356475A444DF3
    claim_text: 'The Ask box''s eval holds cases in English and Polish: answer cases for each of its commands, for questions that need several answerers, for an ambiguous name it must ask about rather than guess and for a question about why the value changed, and refusal cases that ask it to change something, ask whether to buy or sell, or ask about something outside the portfolio'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 365
logic_claims:
  - CLAIM-798356475A444DF3
id: REQ-decision-d219-1-ask-eval-cases
type: req
---
The Ask box's eval holds cases in English and Polish: answer cases for each of its commands, for questions that need several answerers, for an ambiguous name it must ask about rather than guess and for a question about why the value changed, and refusal cases that ask it to change something, ask whether to buy or sell, or ask about something outside the portfolio.

## Context

Author decision D-219 (the Ask box's eval).

## Source

> **[D-219]** The Ask box's eval (D-217) holds cases in English and Polish. Answer cases cover each of its commands (D-194), questions that need several answerers, an ambiguous name it must ask about rather than guess, and a question about why the value changed, every cause of which must come from the cards it shows. Refusal cases ask it to change something, ask whether to buy or sell (R-7), or ask about something outside the portfolio, which it answers with the commands it understands. Prompt injections are planted in an eval-only seed's transaction comments and instrument names and typed by a visitor (R-6). Refusals and injections are its safety cases. A list of English and Polish phrases decides whether an answer gives advice, while a judge model grades the same answers alongside and is only reported until the author lets it decide.
