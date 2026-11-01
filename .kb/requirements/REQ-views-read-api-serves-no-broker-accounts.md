---
title: 'The read API must never serve a broker account: its deprecated ListBrokerAccounts must answer UNIMPLEMENTED over Connect and REST.'
status: closed
text_ref: docs/DESIGN.md#D-121
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-121
  - lane:strict
semantic_text: 'The read API must never serve a broker account: its deprecated ListBrokerAccounts must answer UNIMPLEMENTED over Connect and REST.'
logic_claims:
  - CLAIM-C6E69F00319BF128
semantic_clauses:
  - 'The read API must never serve a broker account: its deprecated ListBrokerAccounts must answer UNIMPLEMENTED over Connect and REST.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0fbb1ddd034b483da16434162f4bc41f14ce5b0b99dabd8309cd8ab72d640214
semantic_inventory:
  - claim_key: CLAIM-C6E69F00319BF128
    claim_text: 'The read API must never serve a broker account: its deprecated ListBrokerAccounts must answer UNIMPLEMENTED over Connect and REST'
    role: normative
    status: modeled
    span:
      start: 0
      end: 129
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:39.119Z'
id: REQ-views-read-api-serves-no-broker-accounts
type: req
---
The read API must never serve a broker account: its deprecated ListBrokerAccounts must answer UNIMPLEMENTED over Connect and REST.

## Context

D-121. buf breaking runs on every change, so the method and its messages stay in the contract, deprecated, rather than being removed (implementation questions). Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> Broker account numbers are listed only through the owner's account service, for setting up accounts. The read API never serves them, so read sessions, script tokens and plugins never see one.

docs/DESIGN.md#D-121
