---
title: The household's grant covers every account, present and future.
status: open
text_ref: design-decisions:D-209#3
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-209-3
semantic_text: The household's grant covers every account, present and future.
logic_claims:
  - CLAIM-F6B646828C7C8520
semantic_clauses:
  - The household's grant covers every account, present and future.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c88e900d245ef10b6aea2ebe1157b3d29bd0302cdda282abe1339a2cd0ee15e6
semantic_inventory:
  - claim_key: CLAIM-F6B646828C7C8520
    claim_text: The household's grant covers every account, present and future
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 62
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-209#3
  recorded_at: '2026-11-13T10:00:34.176Z'
id: REQ-identity-grants-household-covers-every-account
type: req
---
The household's grant covers every account, present and future.

## Context

Part of D-209. The household grant is a reader relationship of the household, which every account belongs to, so an account added later is covered. The decision states no further reason. The author recorded D-209 in DESIGN.md.

## Source

> The owner grants access per account, per sleeve or for the whole household; a sleeve's grant covers the accounts added to it later, and the household's every account, present and future. Viewers, script tokens and view plugins are all granted this way, and a grant only ever lets its holder view. When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-209#3