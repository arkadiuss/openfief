---
title: Broker account numbers must be listed only through the owner's account service, for setting up accounts.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-121-1
  - lane:strict
text_ref: design-decisions:D-121#1
origin:
  kind: import
  ref: design-decisions:D-121#1
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: Broker account numbers must be listed only through the owner's account service, for setting up accounts.
logic_claims:
  - CLAIM-41A5B50FD057D1EA
semantic_clauses:
  - Broker account numbers must be listed only through the owner's account service, for setting up accounts.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 5a8504d66717e0f771c5723fe7f40381781def2772735f475e198628f57e5977
semantic_inventory:
  - claim_key: CLAIM-41A5B50FD057D1EA
    claim_text: Broker account numbers must be listed only through the owner's account service, for setting up accounts
    role: normative
    status: modeled
    span:
      start: 0
      end: 103
id: REQ-ingest-broker-account-numbers-owner-only
type: req
---
Broker account numbers must be listed only through the owner's account service, for setting up accounts.

## Context

D-72 has the UI offer the broker accounts a connector reported, so the owner picks rather than types an account number, and shows the unmapped ones. The implementation listed them through the read API (`ListBrokerAccounts`), which every read caller can call: owner and read sessions, script tokens and view plugins such as Ghostfolio's (question #6 on docs/implementation-questions.md). On 2026-10-31 the author decided that account numbers are listed only through the owner's account service (D-79), which only a view plugin acting for an owner session may call, and that the read API never serves them (REQ-views-read-api-no-broker-account-numbers).

## Source

> Broker account numbers are listed only through the owner's account service, for setting up accounts. The read API never serves them, so read sessions, script tokens and plugins never see one.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-121#1
