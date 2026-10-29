---
title: The owner must create accounts in the built-in UI, in an owner session, giving each account its institution, name, sleeve and the connector instance that feeds it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-72-1
  - lane:strict
text_ref: design-decisions:D-72#1
origin:
  kind: import
  ref: design-decisions:D-72#1
  recorded_at: '2026-10-29T05:37:26.604Z'
semantic_text: The owner must create accounts in the built-in UI, in an owner session, giving each account its institution, name, sleeve and the connector instance that feeds it.
logic_claims:
  - CLAIM-1B340918B79F1F97
semantic_clauses:
  - The owner must create accounts in the built-in UI, in an owner session, giving each account its institution, name, sleeve and the connector instance that feeds it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c7585009d473ed1c20c442b245c7572787da84519ab230734d95a777c6a2ca1a
semantic_inventory:
  - claim_key: CLAIM-1B340918B79F1F97
    claim_text: The owner must create accounts in the built-in UI, in an owner session, giving each account its institution, name, sleeve and the connector instance that feeds it
    role: normative
    status: modeled
    span:
      start: 0
      end: 162
id: REQ-ingest-owner-creates-accounts
type: req
---
The owner must create accounts in the built-in UI, in an owner session, giving each account its institution, name, sleeve and the connector instance that feeds it.

## Context

XTB shut down its public API in 2025, and neither the treasury bond registry nor the bank that sells the bonds offers one, so both arrive as files or manual entry. Interactive Brokers' Flex Queries are expected to report positions per lot, with open date and cost (OQ-22 verifies this). D-72 answers the implementation's question that nothing could create an account outside the demo, because only the demo seed saved accounts and core connector bind needs an account that already exists. The author recorded it with D-65 to D-71 after the design session that answered the implementation's questions.

## Source

> The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-72#1
