---
title: Every broker account a connector instance's delivery reports must be recorded for the instance, with its institution, the day it was last seen and whether the instance's last reporting delivery reported it, before anything of the delivery is imported.
status: open
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
semantic_text: Every broker account a connector instance's delivery reports must be recorded for the instance, with its institution, the day it was last seen and whether the instance's last reporting delivery reported it, before anything of the delivery is imported.
logic_claims:
  - CLAIM-C5AD77EACD3F7E73
semantic_clauses:
  - Every broker account a connector instance's delivery reports must be recorded for the instance, with its institution, the day it was last seen and whether the instance's last reporting delivery reported it, before anything of the delivery is imported.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d9fabe466dd9b1789a95a2b15b1c9af34982979fabb391d327cb89cd4d347b0a
semantic_inventory:
  - claim_key: CLAIM-C5AD77EACD3F7E73
    claim_text: Every broker account a connector instance's delivery reports must be recorded for the instance, with its institution, the day it was last seen and whether the instance's last reporting delivery reported it, before anything of the delivery is imported
    role: normative
    status: modeled
    span:
      start: 0
      end: 250
origin:
  kind: agent
  recorded_at: '2026-10-29T11:36:46.495Z'
id: REQ-ingest-broker-account-record
type: req
---
Every broker account a connector instance's delivery reports must be recorded for the instance, with its institution, the day it was last seen and whether the instance's last reporting delivery reported it, before anything of the delivery is imported.

## Context

D-72: a broker connector reports the broker accounts it saw on its last sync so the UI can offer them; recording them before importing means a refused delivery still shows them. Broker account IDs are account numbers: stored and shown to the owner, never logged. Refines REQ-ingest-broker-account-choice. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-72]** The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

docs/DESIGN.md#D-72
