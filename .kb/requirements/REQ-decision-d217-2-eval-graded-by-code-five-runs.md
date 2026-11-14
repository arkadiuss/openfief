---
title: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run.
status: open
origin:
  kind: import
  ref: design-decisions:D-217#2
  recorded_at: '2026-11-14T08:23:22.151Z'
firmness: firm
semantic_text: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run.
text_ref: design-decisions:D-217#2
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1f8975f117faaebd814837a77ed42c182a085594cd046fdf9f5c684e220a9877
semantic_inventory:
  - claim_key: CLAIM-699BB41E4A59E05C
    claim_text: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 318
logic_claims:
  - CLAIM-699BB41E4A59E05C
id: REQ-decision-d217-2-eval-graded-by-code-five-runs
type: req
---
An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run.

## Context

Author decision D-217 (the eval of a role's model).

## Source

> **[D-217]** A model serves a role only after it passes that role's eval. The eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on `main` to catch a provider changing behind a model's name, each run within a fixed budget. An eval is a fixed set of cases over the demo's synthetic portfolio. Code grades each case against what the read API answers wherever code can decide, and a judge model grades only what code can't, such as whether an answer gives advice. Each case runs five times and passes only if all five pass. Refusal and prompt-injection cases must all pass from the first run; the bar for the rest is set after a baseline run of D-207's two candidates.
