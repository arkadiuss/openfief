---
title: The core must choose a statement upload's mode, opening for an account not yet seeded and transactions after that, with an override for the owner.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-109-3
  - lane:strict
text_ref: design-decisions:D-109#3
origin:
  kind: import
  ref: design-decisions:D-109#3
  recorded_at: '2026-10-29T10:48:18.339Z'
semantic_text: The core must choose a statement upload's mode, opening for an account not yet seeded and transactions after that, with an override for the owner.
logic_claims:
  - CLAIM-4C3CDBE175D18654
semantic_clauses:
  - The core must choose a statement upload's mode, opening for an account not yet seeded and transactions after that, with an override for the owner.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 21bc0327911c984fe984bf39b0da63b943a7df09956f35bda6809d01b8608fa1
semantic_inventory:
  - claim_key: CLAIM-4C3CDBE175D18654
    claim_text: The core must choose a statement upload's mode, opening for an account not yet seeded and transactions after that, with an override for the owner
    role: normative
    status: modeled
    span:
      start: 0
      end: 145
id: REQ-ingest-statement-mode-chosen-by-core
type: req
---
The core must choose a statement upload's mode, opening for an account not yet seeded and transactions after that, with an override for the owner.

## Why this changed

The author's decision D-109 has the core choose an upload's mode, opening for an account not yet seeded and transactions after that (D-95), and lets the owner override it. REQ-ingest-statement-upload-limits refused an upload that came without a mode, so this requirement replaces it; its 64 MiB cap is kept in REQ-ingest-statement-upload-size-limit.

## Context

D-23 seeds accounts from opening lots before transactions are imported, and D-95 has the core switch a seeded account to transactions mode on its own. D-109 applies the same choice to uploads. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> A statement upload goes to the connector instance bound to the account. A statement that covers several accounts, such as an IBKR Flex statement of linked accounts, books every mapped account bound to that instance and reports the broker accounts that aren't mapped (D-72). The core chooses the mode, opening for an account not yet seeded and transactions after that (D-95), and the owner may override it; an opening upload seeds the account for scheduled syncs too. The owner's UI learns through its owner service (D-79) which accounts have a connector that takes files, and shows the upload form only there. The core accepts files up to 64 MiB.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-109#3
