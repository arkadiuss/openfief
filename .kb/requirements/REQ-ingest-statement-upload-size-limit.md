---
title: The core must accept statement files up to 64 MiB.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-109-6
  - lane:strict
text_ref: design-decisions:D-109#6
origin:
  kind: import
  ref: design-decisions:D-109#6
  recorded_at: '2026-10-29T10:50:39.256Z'
semantic_text: The core must accept statement files up to 64 MiB.
logic_claims:
  - CLAIM-E0A068FCB7D19A88
semantic_clauses:
  - The core must accept statement files up to 64 MiB.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 4dd704c41c162aee02c65d674411fae614b9bef67f7e0a2766bd40a0e359d586
semantic_inventory:
  - claim_key: CLAIM-E0A068FCB7D19A88
    claim_text: The core must accept statement files up to 64 MiB
    role: normative
    status: modeled
    span:
      start: 0
      end: 49
id: REQ-ingest-statement-upload-size-limit
type: req
---
The core must accept statement files up to 64 MiB.

## Context

D-109 caps the size of a statement file the core accepts at 64 MiB. Reason not stated. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> A statement upload goes to the connector instance bound to the account. A statement that covers several accounts, such as an IBKR Flex statement of linked accounts, books every mapped account bound to that instance and reports the broker accounts that aren't mapped (D-72). The core chooses the mode, opening for an account not yet seeded and transactions after that (D-95), and the owner may override it; an opening upload seeds the account for scheduled syncs too. The owner's UI learns through its owner service (D-79) which accounts have a connector that takes files, and shows the upload form only there. The core accepts files up to 64 MiB.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-109#6
