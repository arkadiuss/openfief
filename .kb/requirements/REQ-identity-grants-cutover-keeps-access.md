---
title: When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently.
status: open
text_ref: design-decisions:D-209#6
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-209-6
semantic_text: When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently.
logic_claims:
  - CLAIM-D04857F8E610B460
semantic_clauses:
  - When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8091d9f1a5789ff3ac2e56d077567e2b32f5ab9bbdfd9a8868101d1ea20517f6
semantic_inventory:
  - claim_key: CLAIM-D04857F8E610B460
    claim_text: When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently
    role: condition
    status: modeled
    span:
      start: 0
      end: 187
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-209#6
  recorded_at: '2026-11-13T10:00:44.813Z'
id: REQ-identity-grants-cutover-keeps-access
type: req
---
When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently.

## Context

Part of D-209, with the author's answer that the cutover writes one household grant for each existing script token and view plugin, each with its own audit entry, and that the interim policy goes away. The core does it once, at the start in which it first answers from SpiceDB. The Access page that shows these grants is not built yet. The author recorded D-209 in DESIGN.md.

## Source

> The owner grants access per account, per sleeve or for the whole household; a sleeve's grant covers the accounts added to it later, and the household's every account, present and future. Viewers, script tokens and view plugins are all granted this way, and a grant only ever lets its holder view. When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-209#6