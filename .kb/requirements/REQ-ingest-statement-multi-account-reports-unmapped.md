---
title: A statement that covers several accounts must report the broker accounts that aren't mapped.
status: open
text_ref: design-decisions:D-109#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-109-2
  - lane:strict
semantic_text: A statement that covers several accounts must report the broker accounts that aren't mapped.
logic_claims:
  - CLAIM-79A9DD2F95D89D84
semantic_clauses:
  - A statement that covers several accounts must report the broker accounts that aren't mapped.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cbc8c41fa4e55acbd12ac3a0609a7b8dc185368ae58ba6c28b851168ab5bce3d
semantic_inventory:
  - claim_key: CLAIM-79A9DD2F95D89D84
    claim_text: A statement that covers several accounts must report the broker accounts that aren't mapped
    role: normative
    status: modeled
    span:
      start: 0
      end: 91
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-109#2
  recorded_at: '2026-11-01T12:24:18.072Z'
id: REQ-ingest-statement-multi-account-reports-unmapped
type: req
---
A statement that covers several accounts must report the broker accounts that aren't mapped.

## Why this changed

On 2026-11-01 the author recorded D-140, answering the implementation's question #24: an upload of a statement covering several accounts runs in the mode chosen for the account it was uploaded for, so an opening upload seeds only the accounts not seeded yet and a transactions upload books every mapped account. Reason not stated. REQ-ingest-statement-multi-account had such a statement book every mapped account whatever the mode, so this requirement replaces it together with REQ-ingest-multi-account-opening-seeds-unseeded-only and REQ-ingest-multi-account-transactions-books-every-mapped. D-109's report of the broker accounts that aren't mapped is unchanged, and this requirement keeps it.

## Context

D-72 has the UI show a broker account that isn't mapped to an OpenFief account as unmapped, so the owner can map it (REQ-ingest-unmapped-broker-account). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A statement upload goes to the connector instance bound to the account. A statement that covers several accounts, such as an IBKR Flex statement of linked accounts, books every mapped account bound to that instance and reports the broker accounts that aren't mapped (D-72). The core chooses the mode, opening for an account not yet seeded and transactions after that (D-95), and the owner may override it; an opening upload seeds the account for scheduled syncs too. The owner's UI learns through its owner service (D-79) which accounts have a connector that takes files, and shows the upload form only there. The core accepts files up to 64 MiB.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-109#2
