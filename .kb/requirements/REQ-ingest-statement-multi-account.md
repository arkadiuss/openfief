---
title: A statement that covers several accounts must book every mapped account bound to that connector instance and report the broker accounts that aren't mapped.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-109-2
  - lane:strict
text_ref: design-decisions:D-109#2
origin:
  kind: import
  ref: design-decisions:D-109#2
  recorded_at: '2026-10-29T10:47:22.365Z'
semantic_text: A statement that covers several accounts must book every mapped account bound to that connector instance and report the broker accounts that aren't mapped.
logic_claims:
  - CLAIM-5276FFB152B51649
semantic_clauses:
  - A statement that covers several accounts must book every mapped account bound to that connector instance and report the broker accounts that aren't mapped.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 8d83e68c84aa01a740ca0f3179e3a6efcf265036b8d14c43ccd0045a77f0b734
semantic_inventory:
  - claim_key: CLAIM-5276FFB152B51649
    claim_text: A statement that covers several accounts must book every mapped account bound to that connector instance and report the broker accounts that aren't mapped
    role: normative
    status: modeled
    span:
      start: 0
      end: 154
id: REQ-ingest-statement-multi-account
type: req
---
A statement that covers several accounts must book every mapped account bound to that connector instance and report the broker accounts that aren't mapped.

## Why this changed

The author's decision D-109 has a statement that covers several accounts, such as an IBKR Flex statement of linked accounts, book every mapped account bound to the connector instance and report the broker accounts that aren't mapped (D-72). REQ-ingest-statement-upload-handoff confined an upload's batch to the one account it was uploaded for, so this requirement replaces it; handing the file to the bound instance stays in REQ-ingest-statement-upload-bound-instance.

## Context

D-72 has a broker connector report the broker accounts it saw and keeps transactions of an unmapped broker account out of the ledger. D-109 applies that to uploaded statements. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> A statement upload goes to the connector instance bound to the account. A statement that covers several accounts, such as an IBKR Flex statement of linked accounts, books every mapped account bound to that instance and reports the broker accounts that aren't mapped (D-72). The core chooses the mode, opening for an account not yet seeded and transactions after that (D-95), and the owner may override it; an opening upload seeds the account for scheduled syncs too. The owner's UI learns through its owner service (D-79) which accounts have a connector that takes files, and shows the upload form only there. The core accepts files up to 64 MiB.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-109#2
