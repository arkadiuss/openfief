---
title: A grant only ever lets its holder view.
status: open
text_ref: design-decisions:D-209#5
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-209-5
semantic_text: A grant only ever lets its holder view.
logic_claims:
  - CLAIM-D7F5C8FEC2EBD956
semantic_clauses:
  - A grant only ever lets its holder view.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9e532dbb0203673802a32b633814959a61a256ecd1fe445e010aa090d4dccb12
semantic_inventory:
  - claim_key: CLAIM-D7F5C8FEC2EBD956
    claim_text: A grant only ever lets its holder view
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 38
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-209#5
  recorded_at: '2026-11-13T10:00:40.909Z'
id: REQ-identity-grants-only-view
type: req
---
A grant only ever lets its holder view.

## Context

Part of D-209. Only the household's owner may change an account (D-65): the schema gives a grant the viewer and reader relations, which feed the view permission and never the change one. The decision states no further reason. The author recorded D-209 in DESIGN.md.

## Source

> The owner grants access per account, per sleeve or for the whole household; a sleeve's grant covers the accounts added to it later, and the household's every account, present and future. Viewers, script tokens and view plugins are all granted this way, and a grant only ever lets its holder view. When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-209#5