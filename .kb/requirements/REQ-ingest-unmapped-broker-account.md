---
title: Transactions for a broker account that is not mapped to an OpenFief account must not be booked, and the UI must show the broker account as unmapped.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-72-3
  - lane:strict
text_ref: design-decisions:D-72#3
origin:
  kind: import
  ref: design-decisions:D-72#3
  recorded_at: '2026-10-29T05:38:59.733Z'
semantic_text: Transactions for a broker account that is not mapped to an OpenFief account must not be booked, and the UI must show the broker account as unmapped.
logic_claims:
  - CLAIM-E1E448DBCFE98730
semantic_clauses:
  - Transactions for a broker account that is not mapped to an OpenFief account must not be booked, and the UI must show the broker account as unmapped.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 64c18556877e1f7bb332eb18214d1227910c7562f1061e5ae8fbbb8d6142cfd1
semantic_inventory:
  - claim_key: CLAIM-E1E448DBCFE98730
    claim_text: Transactions for a broker account that is not mapped to an OpenFief account must not be booked, and the UI must show the broker account as unmapped
    role: normative
    status: modeled
    span:
      start: 0
      end: 147
id: REQ-ingest-unmapped-broker-account
type: req
---
Transactions for a broker account that is not mapped to an OpenFief account must not be booked, and the UI must show the broker account as unmapped.

## Context

XTB shut down its public API in 2025, and neither the treasury bond registry nor the bank that sells the bonds offers one, so both arrive as files or manual entry. Interactive Brokers' Flex Queries are expected to report positions per lot, with open date and cost (OQ-22 verifies this). The author recorded this decision with D-65 to D-71 after the design session that answered the implementation's questions. REQ-ingest-connector-binding, which refuses a whole delivery that names an account its connector instance is not bound to, is about OpenFief accounts, not broker accounts; how the two meet is not decided yet.

## Source

> The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-72#3
