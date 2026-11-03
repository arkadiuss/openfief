---
title: 'The institutions a connector books for must be a permission its grant approves, as its hostnames are: the signed manifest declares them and the core accepts only what both name.'
status: open
text_ref: design-decisions:D-185#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-185-1
  - lane:strict
semantic_text: 'The institutions a connector books for must be a permission its grant approves, as its hostnames are: the signed manifest declares them and the core accepts only what both name.'
logic_claims:
  - CLAIM-2C2ABA06B41B195F
semantic_clauses:
  - 'The institutions a connector books for must be a permission its grant approves, as its hostnames are: the signed manifest declares them and the core accepts only what both name.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e154dab0577b66c21e2c2657a46618df3ff4457d99934777b2992ebe757d00c0
semantic_inventory:
  - claim_key: CLAIM-2C2ABA06B41B195F
    claim_text: 'The institutions a connector books for must be a permission its grant approves, as its hostnames are: the signed manifest declares them and the core accepts only what both name'
    role: normative
    status: modeled
    span:
      start: 0
      end: 176
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-185#1
  recorded_at: '2026-11-03T07:22:40.646Z'
id: REQ-plugins-institutions-grant-approved
type: req
---
The institutions a connector books for must be a permission its grant approves, as its hostnames are: the signed manifest declares them and the core accepts only what both name.

## Context

The author decided it on 2026-11-03, before the alpha tag, extending D-182's declared institutions: what a connector books for becomes something the owner approves, not only something the vendor declares. D-16 makes every permission a plugin requests something the owner's grant approves. REQ-ingest-delivery-declared-institution (D-182) still holds: an undeclared institution is refused too. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The institutions a connector books for are a permission its grant approves, as its hostnames are (D-16): the signed manifest declares them and the core accepts only what both name. The install warns when another instance already books for an institution, naming that instance. Grants made before this approve the institutions their plugin already declares.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-185#1, decided by the author on 2026-11-03
