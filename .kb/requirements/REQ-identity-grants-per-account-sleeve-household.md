---
title: The owner grants access per account, per sleeve or for the whole household.
status: open
text_ref: design-decisions:D-209#1
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-209-1
semantic_text: The owner grants access per account, per sleeve or for the whole household.
logic_claims:
  - CLAIM-2E2D815839C2D464
semantic_clauses:
  - The owner grants access per account, per sleeve or for the whole household.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ace95538d425e7c1a7edc88094443d6b3bcb5763e4957788cf56aa53aa78e8a0
semantic_inventory:
  - claim_key: CLAIM-2E2D815839C2D464
    claim_text: The owner grants access per account, per sleeve or for the whole household
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 74
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-209#1
  recorded_at: '2026-11-13T10:00:26.539Z'
id: REQ-identity-grants-per-account-sleeve-household
type: req
---
The owner grants access per account, per sleeve or for the whole household.

## Context

The author decided it on 2026-11-12 while planning the Access page and viewers (D-65, D-164, D-210). The identity context stores each grant in a table that is the source of truth, and SpiceDB holds a projection of the grants in force. The owner grants and revokes through the AuthService, in an owner session only. The decision states no further reason. The author recorded D-209 in DESIGN.md.

## Source

> The owner grants access per account, per sleeve or for the whole household; a sleeve's grant covers the accounts added to it later, and the household's every account, present and future. Viewers, script tokens and view plugins are all granted this way, and a grant only ever lets its holder view. When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-209#1