---
title: The core must list broker accounts, with the account each is mapped to, only through the owner's account service, for an owner session.
status: open
text_ref: docs/DESIGN.md#D-121
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-121
  - lane:strict
semantic_text: The core must list broker accounts, with the account each is mapped to, only through the owner's account service, for an owner session.
logic_claims:
  - CLAIM-13C47F0E0B02C18F
semantic_clauses:
  - The core must list broker accounts, with the account each is mapped to, only through the owner's account service, for an owner session.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7113f70db4e9f509c903adabfc4e69ff7b1c5acbd67f28ce61693641047225bc
semantic_inventory:
  - claim_key: CLAIM-13C47F0E0B02C18F
    claim_text: The core must list broker accounts, with the account each is mapped to, only through the owner's account service, for an owner session
    role: normative
    status: modeled
    span:
      start: 0
      end: 134
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:34.513Z'
id: REQ-ingest-broker-accounts-owner-only
type: req
---
The core must list broker accounts, with the account each is mapped to, only through the owner's account service, for an owner session.

## Context

D-121 moves broker account numbers out of the read API, so read sessions, script tokens and plugins never see one. Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> Broker account numbers are listed only through the owner's account service, for setting up accounts. The read API never serves them, so read sessions, script tokens and plugins never see one.

docs/DESIGN.md#D-121
