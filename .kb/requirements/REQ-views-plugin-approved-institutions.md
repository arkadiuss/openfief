---
title: The read API's Plugin must carry the institutions its grant approves next to those its manifest declares, and the UI's plugins page must show the approved ones and say which are not approved.
status: open
text_ref: docs/DESIGN.md#D-185
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-185
  - lane:strict
semantic_text: The read API's Plugin must carry the institutions its grant approves next to those its manifest declares, and the UI's plugins page must show the approved ones and say which are not approved.
logic_claims:
  - CLAIM-B73F126BEAB46723
semantic_clauses:
  - The read API's Plugin must carry the institutions its grant approves next to those its manifest declares, and the UI's plugins page must show the approved ones and say which are not approved.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 90a4d57ce176e67fdf807784ad8f1d3b9ed5570fdd999e3251ce4c5bea114dff
semantic_inventory:
  - claim_key: CLAIM-B73F126BEAB46723
    claim_text: The read API's Plugin must carry the institutions its grant approves next to those its manifest declares, and the UI's plugins page must show the approved ones and say which are not approved
    role: normative
    status: modeled
    span:
      start: 0
      end: 190
origin:
  kind: agent
  recorded_at: '2026-11-03T07:23:17.952Z'
id: REQ-views-plugin-approved-institutions
type: req
---
The read API's Plugin must carry the institutions its grant approves next to those its manifest declares, and the UI's plugins page must show the approved ones and say which are not approved.

## Context

Refines REQ-views-plugin-institutions for D-185: what a connector books for is now what both its manifest and its grant name, so the page shows that and what still waits for the owner. Modeled by Claude Code for the last fixes before the alpha tag, as the orchestrator asked on 2026-11-03.

## Source

> The institutions a connector books for are a permission its grant approves, as its hostnames are (D-16): the signed manifest declares them and the core accepts only what both name. The install warns when another instance already books for an institution, naming that instance. Grants made before this approve the institutions their plugin already declares.

docs/DESIGN.md D-185
