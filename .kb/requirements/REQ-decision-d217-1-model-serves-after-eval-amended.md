---
title: A model serves a role only after it passes that role's eval, and the eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on main, each run within a fixed budget.
status: open
firmness: firm
required_evidence: asserted
semantic_text: A model serves a role only after it passes that role's eval, and the eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on main, each run within a fixed budget.
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
text_ref: design-decisions:D-217#1
origin:
  kind: import
  ref: design-decisions:D-217#1
  recorded_at: '2026-11-14T16:06:13.503Z'
id: REQ-decision-d217-1-model-serves-after-eval-amended
type: req
---
A model serves a role only after it passes that role's eval, and the eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on main, each run within a fixed budget.

## Context

Author decision D-217 (the eval of a role's model), as amended on main: its last clause now ends "the bar for the rest is set after the first baseline run" and no longer names D-207's two candidates, since D-207 now picks the demo's model by the eval. This requirement supersedes REQ-decision-d217-1-model-serves-after-eval, which quotes the old text; the design session approved the supersession on the author's behalf.

## Why this changed

The author amended D-217 and D-207 on main (9c81f8e): the demo's model is now the cheapest one that passes the eval, so D-217 no longer waits for a baseline run of two named candidates; the bar for the answer cases is set after the first baseline run.

## Source

> **[D-217]** A model serves a role only after it passes that role's eval. The eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on `main` to catch a provider changing behind a model's name, each run within a fixed budget. An eval is a fixed set of cases over the demo's synthetic portfolio. Code grades each case against what the read API answers wherever code can decide, and a judge model grades only what code can't, such as whether an answer gives advice. Each case runs five times and passes only if all five pass. Refusal and prompt-injection cases must all pass from the first run; the bar for the rest is set after the first baseline run.
