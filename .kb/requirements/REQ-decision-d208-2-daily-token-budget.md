---
title: The demo's model spend has a daily token budget the core enforces before each call, priced from the model entries, after which the Ask box says the day's demo budget is used up.
status: open
firmness: firm
semantic_text: The demo's model spend has a daily token budget the core enforces before each call, priced from the model entries, after which the Ask box says the day's demo budget is used up.
origin:
  kind: import
  ref: design-decisions:D-208#2
  recorded_at: '2026-11-14T20:37:37.158Z'
text_ref: design-decisions:D-208#2
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 220ea12f48e1f4da4529fd2f90bdfb15cbcabf6527e2af5f5564453bd45b3c0e
semantic_inventory:
  - claim_key: CLAIM-D1FFFC907E0B40C5
    claim_text: The demo's model spend has a daily token budget the core enforces before each call, priced from the model entries, after which the Ask box says the day's demo budget is used up
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 176
    subjects:
      - models.router_daily_budget
logic_claims:
  - CLAIM-D1FFFC907E0B40C5
id: REQ-decision-d208-2-daily-token-budget
type: req
pair_reviews:
  - with: REQ-models-daily-budget
    verdict: no_conflict
    separated_by: the demo's daily token budget and what the Ask box says when it is used up (D-208) against the router's counting of every call's cost in dollars and its refusal of a call before it is sent; priced tokens are dollars, so they state one limit from two sides and the second implements the first
    claims:
      - claim_key: CLAIM-004DBD465FC462EC
      - claim_key: CLAIM-88774B46B9233BDA
      - claim_key: CLAIM-D1FFFC907E0B40C5
      - claim_key: CLAIM-F3841463A5ABFE49
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
---
The demo's model spend has a daily token budget the core enforces before each call, priced from the model entries, after which the Ask box says the day's demo budget is used up.

## Context

Author decision D-208 (the demo's spend guards), imported from docs/DESIGN.md; it was not imported before. REQ-models-daily-budget is the behavior: the cap, the pricing from the entries, the refusal before the call and the Ask box's card.

## Source

> docs/DESIGN.md, D-208#2.