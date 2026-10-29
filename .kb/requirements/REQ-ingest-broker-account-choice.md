---
title: A broker connector must report the broker accounts it saw on its last sync, and the UI must offer them to pick from, so that the owner never types a broker account number.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-72-2
  - lane:strict
text_ref: design-decisions:D-72#2
origin:
  kind: import
  ref: design-decisions:D-72#2
  recorded_at: '2026-10-29T05:38:14.982Z'
semantic_text: A broker connector must report the broker accounts it saw on its last sync, and the UI must offer them to pick from, so that the owner never types a broker account number.
logic_claims:
  - CLAIM-73CB3E1460FB804A
semantic_clauses:
  - A broker connector must report the broker accounts it saw on its last sync, and the UI must offer them to pick from, so that the owner never types a broker account number.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 174468c38607181345728efaa4cdfbc944c480ef3e0f03c173e10313c1a644c2
semantic_inventory:
  - claim_key: CLAIM-73CB3E1460FB804A
    claim_text: A broker connector must report the broker accounts it saw on its last sync, and the UI must offer them to pick from, so that the owner never types a broker account number
    role: normative
    status: modeled
    span:
      start: 0
      end: 170
id: REQ-ingest-broker-account-choice
type: req
---
A broker connector must report the broker accounts it saw on its last sync, and the UI must offer them to pick from, so that the owner never types a broker account number.

## Context

XTB shut down its public API in 2025, and neither the treasury bond registry nor the bank that sells the bonds offers one, so both arrive as files or manual entry. Interactive Brokers' Flex Queries are expected to report positions per lot, with open date and cost (OQ-22 verifies this). The author recorded this decision with D-65 to D-71 after the design session that answered the implementation's questions; the owner picks rather than types, so no broker account number is entered by hand.

## Source

> The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-72#2
