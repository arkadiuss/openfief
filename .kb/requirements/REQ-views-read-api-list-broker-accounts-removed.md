---
title: The read API's ListBrokerAccounts method and its messages must be removed from the contract, not kept as deprecated.
status: open
text_ref: design-decisions:D-157#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-157-1
  - lane:strict
semantic_text: The read API's ListBrokerAccounts method and its messages must be removed from the contract, not kept as deprecated.
logic_claims:
  - CLAIM-04E3AC37CEE96A18
semantic_clauses:
  - The read API's ListBrokerAccounts method and its messages must be removed from the contract, not kept as deprecated.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 37edb4d890a271f2fcba32eef2015a59777aa85ffa6be7720dfa933fca73a720
semantic_inventory:
  - claim_key: CLAIM-04E3AC37CEE96A18
    claim_text: The read API's ListBrokerAccounts method and its messages must be removed from the contract, not kept as deprecated
    role: normative
    status: modeled
    span:
      start: 0
      end: 115
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-157#1
  recorded_at: '2026-11-01T12:30:51.623Z'
id: REQ-views-read-api-list-broker-accounts-removed
type: req
---
The read API's ListBrokerAccounts method and its messages must be removed from the contract, not kept as deprecated.

## Context

The author recorded it on 2026-11-01, settling the last implementation questions (#33, #38, #39). D-121 lists broker account numbers only through the owner's account service, and the read API never serves them (REQ-views-read-api-no-broker-account-numbers). No end-to-end test proves it yet, so its evidence floor is asserted.

## Why this changed

It supersedes REQ-views-read-api-serves-no-broker-accounts, which kept ListBrokerAccounts in the contract as a deprecated method answering UNIMPLEMENTED so the read API changed only by addition. D-157 removes the method and its messages instead, as the one breaking change the author approved.

## Source

> The read API's `ListBrokerAccounts` method and its messages are removed from the contract, not kept as deprecated. The breaking-change check lets this one removal through as an exception the author approved, and stays on for every other change.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-157#1
