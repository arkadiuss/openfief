---
title: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run, the bar for the rest being set after the first baseline run.
status: open
firmness: firm
required_evidence: asserted
semantic_text: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run, the bar for the rest being set after the first baseline run.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e7b2ead24b8bb650ab2d039a1cb1ad168c21899c03e08a56b57bcc2ed4afaeca
semantic_inventory:
  - claim_key: CLAIM-10E2B89956DC0076
    claim_text: An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run, the bar for the rest being set after the first baseline run
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 379
logic_claims:
  - CLAIM-10E2B89956DC0076
text_ref: design-decisions:D-217#2
origin:
  kind: import
  ref: design-decisions:D-217#2
  recorded_at: '2026-11-14T16:06:27.002Z'
id: REQ-decision-d217-2-eval-graded-by-code-five-runs-amended
type: req
---
An eval is a fixed set of cases over the demo's synthetic portfolio, graded by code against what the read API answers wherever code can decide and by a judge model only where code can't, each case runs five times and passes only if all five pass, and refusal and prompt-injection cases must all pass from the first run, the bar for the rest being set after the first baseline run.

## Context

Author decision D-217 (the eval of a role's model), as amended on main: its last clause now ends "the bar for the rest is set after the first baseline run" and no longer names D-207's two candidates, since D-207 now picks the demo's model by the eval. This requirement supersedes REQ-decision-d217-2-eval-graded-by-code-five-runs, which quotes the old text; the design session approved the supersession on the author's behalf.

## Why this changed

The author amended D-217 and D-207 on main (9c81f8e): the demo's model is now the cheapest one that passes the eval, so D-217 no longer waits for a baseline run of two named candidates; the bar for the answer cases is set after the first baseline run.

## Source

> **[D-217]** A model serves a role only after it passes that role's eval. The eval runs again whenever the role's model, provider host, prompt or tools change, and weekly on `main` to catch a provider changing behind a model's name, each run within a fixed budget. An eval is a fixed set of cases over the demo's synthetic portfolio. Code grades each case against what the read API answers wherever code can decide, and a judge model grades only what code can't, such as whether an answer gives advice. Each case runs five times and passes only if all five pass. Refusal and prompt-injection cases must all pass from the first run; the bar for the rest is set after the first baseline run.
