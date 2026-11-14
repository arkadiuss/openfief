---
title: A model serves a role only after it passes that role's eval, and the eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on main, each run within a fixed budget.
status: open
origin:
  kind: import
  ref: design-decisions:D-217#1
  recorded_at: '2026-11-14T08:23:19.231Z'
firmness: firm
semantic_text: A model serves a role only after it passes that role's eval, and the eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on main, each run within a fixed budget.
text_ref: design-decisions:D-217#1
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 011ef9727f7be3bd9b32f6c1957dacd97d1bb7421dec06001c4dde280e579ab1
semantic_inventory:
  - claim_key: CLAIM-6E4D5E0125451A64
    claim_text: A model serves a role only after it passes that role's eval, and the eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on main, each run within a fixed budget
    role: descriptive
    status: missing
    span:
      start: 0
      end: 201
logic_claims:
  - CLAIM-6E4D5E0125451A64
id: REQ-decision-d217-1-model-serves-after-eval
type: req
---
A model serves a role only after it passes that role's eval, and the eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on main, each run within a fixed budget.

## Context

Author decision D-217 (the eval of a role's model).

## Source

> **[D-217]** A model serves a role only after it passes that role's eval. The eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on `main` to catch a provider changing behind a model's name, each run within a fixed budget. An eval is a fixed set of cases over the demo's synthetic portfolio. Code grades each case against what the read API answers wherever code can decide, and a judge model grades only what code can't, such as whether an answer gives advice. Each case runs five times and passes only if all five pass. Refusal and prompt-injection cases must all pass from the first run; the bar for the rest is set after a baseline run of D-207's two candidates.
