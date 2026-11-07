---
title: A call to a connector from any pod other than the core must fail, and a call from the core without a valid token for that instance must be refused.
status: open
text_ref: design-decisions:D-200#17
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-17
  - lane:strict
semantic_text: A call to a connector from any pod other than the core must fail, and a call from the core without a valid token for that instance must be refused.
logic_claims:
  - CLAIM-40BF7FAD4973F58F
semantic_clauses:
  - A call to a connector from any pod other than the core must fail, and a call from the core without a valid token for that instance must be refused.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 60c57dd9790c1b1a31875366b2a4432955860353dcaca99ce6fda8faa7b9753b
semantic_inventory:
  - claim_key: CLAIM-40BF7FAD4973F58F
    claim_text: A call to a connector from any pod other than the core must fail, and a call from the core without a valid token for that instance must be refused
    role: normative
    status: modeled
    span:
      start: 0
      end: 146
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#17
  recorded_at: '2026-11-07T03:32:45.179Z'
id: REQ-m0-exit-17-core-to-plugin
type: req
---
A call to a connector from any pod other than the core must fail, and a call from the core without a valid token for that instance must be refused.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-17.

## Source

> **Core-to-plugin.** A call to a connector from any pod other than the core fails, and a call from the core without a valid token for that instance is refused.

design-decisions:D-200#17 (docs/DESIGN.md, D-200, criterion 17)

## Why this changed

An earlier agent modeled criterion 17 as the default-firmness requirement REQ-e2e-core-plugin-tokens before D-200's criteria were imported as firm requirements. D-200 is the author's decision, so this firm requirement, stated from criterion 17 itself, replaces it, and the end-to-end scenarios that verified REQ-e2e-core-plugin-tokens now specify this requirement too. Its last clause, that the core's plugin port refuses a plugin's call without a valid view plugin token, goes beyond criterion 17 and stays as the default-firmness refinement REQ-e2e-core-plugin-port-refuses-bad-tokens.
