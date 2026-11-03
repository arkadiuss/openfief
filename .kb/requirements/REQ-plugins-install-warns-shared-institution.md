---
title: The install must warn when another instance already books for an institution, naming that instance.
status: open
text_ref: design-decisions:D-185#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-185-2
  - lane:strict
semantic_text: The install must warn when another instance already books for an institution, naming that instance.
logic_claims:
  - CLAIM-5652DC0075E89F7A
semantic_clauses:
  - The install must warn when another instance already books for an institution, naming that instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 561b9d048c5770f2e3b1370fde2729cfb7bfe5a2110f13448d5bdbe54ad27e24
semantic_inventory:
  - claim_key: CLAIM-5652DC0075E89F7A
    claim_text: The install must warn when another instance already books for an institution, naming that instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 98
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-185#2
  recorded_at: '2026-11-03T07:22:48.963Z'
id: REQ-plugins-install-warns-shared-institution
type: req
---
The install must warn when another instance already books for an institution, naming that instance.

## Context

The author decided it on 2026-11-03, before the alpha tag, extending D-182's declared institutions: what a connector books for becomes something the owner approves, not only something the vendor declares. Two instances may book for one institution, such as two accounts at the same broker; the warning lets the owner see it before confirming. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The institutions a connector books for are a permission its grant approves, as its hostnames are (D-16): the signed manifest declares them and the core accepts only what both name. The install warns when another instance already books for an institution, naming that instance. Grants made before this approve the institutions their plugin already declares.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-185#2, decided by the author on 2026-11-03
