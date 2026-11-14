---
title: Prompt injections are planted in an eval-only seed's transaction comments and instrument names and typed by a visitor, and never in the public demo seed.
status: open
origin:
  kind: import
  ref: design-decisions:D-219#2
  recorded_at: '2026-11-14T08:23:30.524Z'
firmness: firm
semantic_text: Prompt injections are planted in an eval-only seed's transaction comments and instrument names and typed by a visitor, and never in the public demo seed.
text_ref: design-decisions:D-219#2
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6a120ea6e7ecfe934b561805c81e94a3f07534e3122cc9e5d9ef7b6b7abcee1b
semantic_inventory:
  - claim_key: CLAIM-361D335E90EA7457
    claim_text: Prompt injections are planted in an eval-only seed's transaction comments and instrument names and typed by a visitor, and never in the public demo seed
    role: descriptive
    status: missing
    span:
      start: 0
      end: 152
logic_claims:
  - CLAIM-361D335E90EA7457
id: REQ-decision-d219-2-injections-in-eval-only-seed
type: req
---
Prompt injections are planted in an eval-only seed's transaction comments and instrument names and typed by a visitor, and never in the public demo seed.

## Context

Author decision D-219 (the Ask box's eval).

## Source

> **[D-219]** The Ask box's eval (D-217) holds cases in English and Polish. Answer cases cover each of its commands (D-194), questions that need several answerers, an ambiguous name it must ask about rather than guess, and a question about why the value changed, every cause of which must come from the cards it shows. Refusal cases ask it to change something, ask whether to buy or sell (R-7), or ask about something outside the portfolio, which it answers with the commands it understands. Prompt injections are planted in an eval-only seed's transaction comments and instrument names and typed by a visitor (R-6). Refusals and injections are its safety cases. A list of English and Polish phrases decides whether an answer gives advice, while a judge model grades the same answers alongside and is only reported until the author lets it decide.
