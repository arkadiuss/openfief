---
title: The owner must be able to remap a broker account while its OpenFief account holds no version of a transaction from the connector instance that reports it, and not otherwise.
status: open
text_ref: design-decisions:D-156#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-156-1
  - lane:strict
semantic_text: The owner must be able to remap a broker account while its OpenFief account holds no version of a transaction from the connector instance that reports it, and not otherwise.
logic_claims:
  - CLAIM-88E6CC19AAE4EDB0
semantic_clauses:
  - The owner must be able to remap a broker account while its OpenFief account holds no version of a transaction from the connector instance that reports it, and not otherwise.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6c691dc8299bdd6c88148e025988eab9ae63eca8c31692609d75ec859ad90bcf
semantic_inventory:
  - claim_key: CLAIM-88E6CC19AAE4EDB0
    claim_text: The owner must be able to remap a broker account while its OpenFief account holds no version of a transaction from the connector instance that reports it, and not otherwise
    role: normative
    status: modeled
    span:
      start: 0
      end: 172
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-156#1
  recorded_at: '2026-11-01T12:30:25.381Z'
id: REQ-ingest-broker-account-remap-per-instance
type: req
---
The owner must be able to remap a broker account while its OpenFief account holds no version of a transaction from the connector instance that reports it, and not otherwise.

## Why this changed

On 2026-11-01 the author recorded D-156, settling how D-120's remap check counts transactions. The decision gives the reason: the ledger doesn't record which broker account a transaction came from, so the check looks at the connector instance that reports the broker account; when one instance feeds several broker accounts into the same account it refuses some remaps it could allow, and never allows an unsafe one. REQ-ingest-broker-account-remap promised a remap whenever the account had no transactions from that broker account, which the per-instance check no longer allows in every case, so this requirement replaces it.

## Context

D-120 lets the owner remap a broker account whose OpenFief account has no transactions from it, and D-156 says what counts as having none. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A broker account counts as having no transactions in its OpenFief account (D-120) while that account holds no version of a transaction from the connector instance that reports it; balance assertions don't count. The ledger doesn't record which broker account a transaction came from, so when one instance feeds several broker accounts into the same account the check refuses some remaps it could allow, and never allows an unsafe one.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-156#1
