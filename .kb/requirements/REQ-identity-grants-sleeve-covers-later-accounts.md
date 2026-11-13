---
title: A sleeve's grant covers the accounts added to it later.
status: open
text_ref: design-decisions:D-209#2
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-209-2
semantic_text: A sleeve's grant covers the accounts added to it later.
logic_claims:
  - CLAIM-74ABE339B3472781
semantic_clauses:
  - A sleeve's grant covers the accounts added to it later.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 07c98bead3826f31ace139f86de0bb8c3127caeaeecf6bf91c2d211ea3dfa232
semantic_inventory:
  - claim_key: CLAIM-74ABE339B3472781
    claim_text: A sleeve's grant covers the accounts added to it later
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 54
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-209#2
  recorded_at: '2026-11-13T10:00:30.376Z'
id: REQ-identity-grants-sleeve-covers-later-accounts
type: req
---
A sleeve's grant covers the accounts added to it later.

## Context

Part of D-209. A grant names the sleeve, never a list of accounts, and the authorization model gives an account its sleeve's viewers, so an account placed in the sleeve afterwards is covered with no new grant. The decision states no further reason. The author recorded D-209 in DESIGN.md.

## Source

> The owner grants access per account, per sleeve or for the whole household; a sleeve's grant covers the accounts added to it later, and the household's every account, present and future. Viewers, script tokens and view plugins are all granted this way, and a grant only ever lets its holder view. When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-209#2