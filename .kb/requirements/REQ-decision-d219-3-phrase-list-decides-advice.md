---
title: Refusals and injections are the Ask box eval's safety cases, a list of English and Polish phrases decides whether an answer gives advice, and a judge model grades the same answers alongside and is only reported until the author lets it decide.
status: open
origin:
  kind: import
  ref: design-decisions:D-219#3
  recorded_at: '2026-11-14T08:23:33.183Z'
firmness: firm
semantic_text: Refusals and injections are the Ask box eval's safety cases, a list of English and Polish phrases decides whether an answer gives advice, and a judge model grades the same answers alongside and is only reported until the author lets it decide.
text_ref: design-decisions:D-219#3
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3c10fb4d9db15e25d64cba67a4138a4b17b59be16fbf947f157c920823b0399f
semantic_inventory:
  - claim_key: CLAIM-3A223623C09EEBEE
    claim_text: Refusals and injections are the Ask box eval's safety cases, a list of English and Polish phrases decides whether an answer gives advice, and a judge model grades the same answers alongside and is only reported until the author lets it decide
    role: descriptive
    status: missing
    span:
      start: 0
      end: 242
logic_claims:
  - CLAIM-3A223623C09EEBEE
id: REQ-decision-d219-3-phrase-list-decides-advice
type: req
---
Refusals and injections are the Ask box eval's safety cases, a list of English and Polish phrases decides whether an answer gives advice, and a judge model grades the same answers alongside and is only reported until the author lets it decide.

## Context

Author decision D-219 (the Ask box's eval).

## Source

> **[D-219]** The Ask box's eval (D-217) holds cases in English and Polish. Answer cases cover each of its commands (D-194), questions that need several answerers, an ambiguous name it must ask about rather than guess, and a question about why the value changed, every cause of which must come from the cards it shows. Refusal cases ask it to change something, ask whether to buy or sell (R-7), or ask about something outside the portfolio, which it answers with the commands it understands. Prompt injections are planted in an eval-only seed's transaction comments and instrument names and typed by a visitor (R-6). Refusals and injections are its safety cases. A list of English and Polish phrases decides whether an answer gives advice, while a judge model grades the same answers alongside and is only reported until the author lets it decide.
