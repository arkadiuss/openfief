---
title: The built-in UI must take broker accounts from the owner's account service and show them to an owner session only.
status: open
text_ref: docs/DESIGN.md#D-121
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-121
  - lane:strict
semantic_text: The built-in UI must take broker accounts from the owner's account service and show them to an owner session only.
logic_claims:
  - CLAIM-0C2A4B98FC4C408A
semantic_clauses:
  - The built-in UI must take broker accounts from the owner's account service and show them to an owner session only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0da58e009677679286c98c40abdaac61222d389c3e11bdc53662ca0225980a10
semantic_inventory:
  - claim_key: CLAIM-0C2A4B98FC4C408A
    claim_text: The built-in UI must take broker accounts from the owner's account service and show them to an owner session only
    role: normative
    status: modeled
    span:
      start: 0
      end: 113
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:50.527Z'
id: REQ-ui-broker-accounts-owner-only
type: req
---
The built-in UI must take broker accounts from the owner's account service and show them to an owner session only.

## Context

D-121: a read session, such as the demo's, never sees a broker account number. Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> Broker account numbers are listed only through the owner's account service, for setting up accounts. The read API never serves them, so read sessions, script tokens and plugins never see one.

docs/DESIGN.md#D-121
