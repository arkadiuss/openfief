---
title: The read API must never serve broker account numbers, so that read sessions, script tokens and plugins never see one.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-121-2
  - lane:strict
text_ref: design-decisions:D-121#2
origin:
  kind: import
  ref: design-decisions:D-121#2
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: The read API must never serve broker account numbers, so that read sessions, script tokens and plugins never see one.
logic_claims:
  - CLAIM-1B6C42E30707A03F
semantic_clauses:
  - The read API must never serve broker account numbers, so that read sessions, script tokens and plugins never see one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 54671bec1c77ae58dd9f3d0c8f2a3aa02e78b04527594859f913900b6d4a5219
semantic_inventory:
  - claim_key: CLAIM-1B6C42E30707A03F
    claim_text: The read API must never serve broker account numbers, so that read sessions, script tokens and plugins never see one
    role: normative
    status: modeled
    span:
      start: 0
      end: 116
id: REQ-views-read-api-no-broker-account-numbers
type: req
---
The read API must never serve broker account numbers, so that read sessions, script tokens and plugins never see one.

## Context

A broker account ID is the broker's account number. The read API listed the broker accounts connector instances reported (`ListBrokerAccounts`), and every read caller could call it: owner and read sessions, script tokens and view plugins such as Ghostfolio's (question #6 on docs/implementation-questions.md). On 2026-10-31 the author decided that those numbers are listed only through the owner's account service (REQ-ingest-broker-account-numbers-owner-only) and that the read API never serves them, so only an owner session setting up accounts ever sees one.

## Source

> Broker account numbers are listed only through the owner's account service, for setting up accounts. The read API never serves them, so read sessions, script tokens and plugins never see one.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-121#2
