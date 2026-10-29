---
title: An opening upload must seed the account for scheduled syncs too.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-109-4
  - lane:strict
text_ref: design-decisions:D-109#4
origin:
  kind: import
  ref: design-decisions:D-109#4
  recorded_at: '2026-10-29T10:49:06.086Z'
semantic_text: An opening upload must seed the account for scheduled syncs too.
logic_claims:
  - CLAIM-18930CE390EE7998
semantic_clauses:
  - An opening upload must seed the account for scheduled syncs too.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2d52a963d5da0fe4158b955648fd1a429ed82feb5eecba1c7d7bb88759e4cb15
semantic_inventory:
  - claim_key: CLAIM-18930CE390EE7998
    claim_text: An opening upload must seed the account for scheduled syncs too
    role: normative
    status: modeled
    span:
      start: 0
      end: 63
id: REQ-ingest-opening-upload-seeds-syncs
type: req
---
An opening upload must seed the account for scheduled syncs too.

## Context

D-95 has the core sync an account already seeded from opening lots in transactions mode on its own (REQ-ingest-seeded-account-transactions-mode). D-109 makes an opening upload count as that seeding, so the scheduled syncs that follow it run in transactions mode. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> A statement upload goes to the connector instance bound to the account. A statement that covers several accounts, such as an IBKR Flex statement of linked accounts, books every mapped account bound to that instance and reports the broker accounts that aren't mapped (D-72). The core chooses the mode, opening for an account not yet seeded and transactions after that (D-95), and the owner may override it; an opening upload seeds the account for scheduled syncs too. The owner's UI learns through its owner service (D-79) which accounts have a connector that takes files, and shows the upload form only there. The core accepts files up to 64 MiB.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-109#4
