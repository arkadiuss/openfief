---
title: The owner must be able to remap a broker account whose OpenFief account has no transactions from it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-120-2
  - lane:strict
text_ref: design-decisions:D-120#2
origin:
  kind: import
  ref: design-decisions:D-120#2
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: The owner must be able to remap a broker account whose OpenFief account has no transactions from it.
logic_claims:
  - CLAIM-A0BB039E8C08654A
semantic_clauses:
  - The owner must be able to remap a broker account whose OpenFief account has no transactions from it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 8d20609f13a171d11645062ab6a8ffff9f965b1c643b3b25819ce4840d6527cc
semantic_inventory:
  - claim_key: CLAIM-A0BB039E8C08654A
    claim_text: The owner must be able to remap a broker account whose OpenFief account has no transactions from it
    role: normative
    status: modeled
    span:
      start: 0
      end: 99
id: REQ-ingest-broker-account-remap
type: req
---
The owner must be able to remap a broker account whose OpenFief account has no transactions from it.

## Context

D-72 has the owner map each broker account a connector reports to an OpenFief account (REQ-ingest-broker-account-choice), but didn't say whether a mapping can change, so a mapped broker account stayed mapped and mapping it elsewhere was refused (question #5 on docs/implementation-questions.md). On 2026-10-31 the author allowed remapping while the mapped OpenFief account holds no transactions from that broker account, because nothing the ledger holds is ever deleted or moved (D-18).

## Source

> The owner can delete a sleeve with no accounts, a sleeve kind no sleeve uses and an account with no transactions, and remap a broker account whose OpenFief account has no transactions from it. Nothing the ledger holds is ever deleted (D-18).

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-120#2
