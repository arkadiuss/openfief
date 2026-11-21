---
title: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run, each role's bar for the rest being its own decision, D-220 for the Ask box.
status: closed
firmness: firm
required_evidence: asserted
semantic_text: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run, each role's bar for the rest being its own decision, D-220 for the Ask box.
origin:
  kind: import
  ref: design-decisions:D-217#2
  recorded_at: '2026-11-14T18:40:16.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7ce1039d383ed12a1eba6eccf75cbb13862d39557623b23bb6ea28e0d423cf1a
semantic_inventory:
  - claim_key: CLAIM-5B3D41C8323606A2
    claim_text: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run, each role's bar for the rest being its own decision, D-220 for the Ask box
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 394
    subjects:
      - ask_eval.grading_and_runs
logic_claims:
  - CLAIM-5B3D41C8323606A2
text_ref: design-decisions:D-217#2
id: REQ-decision-d217-2-eval-graded-by-code-five-runs-d220
type: req
---
An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run, each role's bar for the rest being its own decision, D-220 for the Ask box.

## Context

Author decision D-217 (the eval of a role's model), as amended on main by D-220: its last clause now ends "each role's bar for the rest is its own decision, D-220 for the Ask box" and no longer waits for a baseline run. This requirement supersedes REQ-decision-d217-2-eval-graded-by-code-five-runs-amended, which quotes the earlier text; the design session approved the supersession on the author's behalf.

## Why this changed

The author amended D-217 and added D-220: the bar for the Ask box's cases that are not safety cases is now stated, so it is no longer set after a baseline run.

## Source

> **[D-217]** A model serves a role only after it passes that role's eval. The eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on `main` to catch a provider changing behind a model's name, each run within a fixed budget. An eval is a fixed set of cases over the demo's synthetic portfolio. Code grades each case against what the read API answers wherever code can decide, and a judge model grades only what code can't, such as whether an answer gives advice. Each case runs five times and passes only if all five pass. Refusal and prompt-injection cases must all pass from the first run; each role's bar for the rest is its own decision, D-220 for the Ask box.
