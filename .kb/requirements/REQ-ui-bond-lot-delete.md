---
title: The built-in UI must offer the owner a Delete button for each open lot on the page of an account at the treasury bond registry, hand the lot's series and purchase date to the core's BondLotService, and show the core's reason when it is refused.
status: open
text_ref: docs/DESIGN.md#D-100
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-100
  - lane:strict
semantic_text: The built-in UI must offer the owner a Delete button for each open lot on the page of an account at the treasury bond registry, hand the lot's series and purchase date to the core's BondLotService, and show the core's reason when it is refused.
logic_claims:
  - CLAIM-810B024E0CCEBC57
semantic_clauses:
  - The built-in UI must offer the owner a Delete button for each open lot on the page of an account at the treasury bond registry, hand the lot's series and purchase date to the core's BondLotService, and show the core's reason when it is refused.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5651004c3d341028f5593667e0558101d8c5225efb6287a318e58107f35a61bd
semantic_inventory:
  - claim_key: CLAIM-810B024E0CCEBC57
    claim_text: The built-in UI must offer the owner a Delete button for each open lot on the page of an account at the treasury bond registry, hand the lot's series and purchase date to the core's BondLotService, and show the core's reason when it is refused
    role: normative
    status: modeled
    span:
      start: 0
      end: 243
origin:
  kind: agent
  recorded_at: '2026-10-30T07:37:25.554Z'
id: REQ-ui-bond-lot-delete
type: req
---
The built-in UI must offer the owner a Delete button for each open lot on the page of an account at the treasury bond registry, hand the lot's series and purchase date to the core's BondLotService, and show the core's reason when it is refused.

## Context

D-100 lets the owner delete a lot as a correction; D-79 sends what the owner enters through the core's owner service. A button per open lot is this package's narrow reading of the UI. Recorded by the bonds implementation agent (impl/bonds-2).

## Source

> Manual bond lots are cash purchases at 100 PLN or exchanges of maturing bonds at 99.90 PLN, whose 0.10 PLN discount is taxed with the interest on redemption. The owner can delete a lot as a correction, which supersedes it (D-18) and erases nothing.

docs/DESIGN.md#D-100
