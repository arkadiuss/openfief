---
title: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't. Each case runs five times. Refusal and prompt-injection cases must pass all five runs from the first run; each role's bar for the rest is its own decision, D-220 for the Ask box.
status: open
firmness: firm
required_evidence: asserted
semantic_text: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't. Each case runs five times. Refusal and prompt-injection cases must pass all five runs from the first run; each role's bar for the rest is its own decision, D-220 for the Ask box.
text_ref: design-decisions:D-217#2
origin:
  kind: import
  ref: design-decisions:D-217#2
  recorded_at: '2026-11-20T22:42:20.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dedb0734b7f158a36fd11ea973f408d3a555eb041ca8467f69a2beb0992bb7a1
semantic_inventory:
  - claim_key: CLAIM-AE7F0397CEECCA25
    claim_text: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 185
    subjects:
      - ask_eval.grading_and_runs
  - claim_key: CLAIM-9459A4CDFAA22123
    claim_text: Each case runs five times
    role: descriptive
    status: missing
    span:
      start: 187
      end: 212
  - claim_key: CLAIM-739218517389FDB2
    claim_text: Refusal and prompt-injection cases must pass all five runs from the first run
    role: normative
    status: ontology_gap
    span:
      start: 214
      end: 291
    subjects:
      - ask_eval.grading_and_runs
  - claim_key: CLAIM-3C3AE2E228A7CB2E
    claim_text: each role's bar for the rest is its own decision, D-220 for the Ask box
    role: descriptive
    status: missing
    span:
      start: 293
      end: 364
logic_claims:
  - CLAIM-AE7F0397CEECCA25
  - CLAIM-9459A4CDFAA22123
  - CLAIM-739218517389FDB2
  - CLAIM-3C3AE2E228A7CB2E
id: REQ-decision-d217-2-eval-graded-by-code-five-runs-each-roles-bar
type: req
---
An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't. Each case runs five times. Refusal and prompt-injection cases must pass all five runs from the first run; each role's bar for the rest is its own decision, D-220 for the Ask box.

## Context

Author decision D-217 (the eval of a role's model), as the author rewrote its last two sentences on 2026-11-20: "Each case runs five times" no longer says a case passes only if all five pass, and the refusal and prompt-injection cases are the ones that must pass all five runs from the first run; each role's bar for the rest is its own decision, D-220 for the Ask box. This requirement supersedes REQ-decision-d217-2-eval-graded-by-code-five-runs-d220, which quotes the earlier text. The scenario that specified it moves here.

## Why this changed

The author changed D-217 so that the pass rule is stated for the safety cases and left to each role's bar for every other case, instead of stating one rule (all five runs) for every case; the Ask box's own bar is D-220. The design session relayed the decision.

## Source

> **[D-217]** ... Each case runs five times. Refusal and prompt-injection cases must pass all five runs from the first run; each role's bar for the rest is its own decision, D-220 for the Ask box.
