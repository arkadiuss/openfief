---
title: The breaking-change check must let the removal of ListBrokerAccounts through as the one exception the author approved, and stay on for every other change.
status: open
text_ref: design-decisions:D-157#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-157-2
  - lane:strict
semantic_text: The breaking-change check must let the removal of ListBrokerAccounts through as the one exception the author approved, and stay on for every other change.
logic_claims:
  - CLAIM-1233843E333AB5C6
semantic_clauses:
  - The breaking-change check must let the removal of ListBrokerAccounts through as the one exception the author approved, and stay on for every other change.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3f421be75a5782ef11b37c346e04c4f40ccdc9abf96fa3bfb41e16c7659d9295
semantic_inventory:
  - claim_key: CLAIM-1233843E333AB5C6
    claim_text: The breaking-change check must let the removal of ListBrokerAccounts through as the one exception the author approved, and stay on for every other change
    role: normative
    status: modeled
    span:
      start: 0
      end: 153
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-157#2
  recorded_at: '2026-11-01T12:31:04.108Z'
id: REQ-tooling-breaking-check-one-exception
type: req
---
The breaking-change check must let the removal of ListBrokerAccounts through as the one exception the author approved, and stay on for every other change.

## Context

The author recorded it on 2026-11-01, settling the last implementation questions (#33, #38, #39). The design session's commit gives the reason: a breaking contract change needs an author decision naming it. The contracts are checked with buf breaking against the base revision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The read API's `ListBrokerAccounts` method and its messages are removed from the contract, not kept as deprecated. The breaking-change check lets this one removal through as an exception the author approved, and stays on for every other change.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-157#2
