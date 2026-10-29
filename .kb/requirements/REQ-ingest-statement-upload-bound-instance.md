---
title: A statement upload must go to the connector instance bound to the account.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-109-1
  - lane:strict
text_ref: design-decisions:D-109#1
origin:
  kind: import
  ref: design-decisions:D-109#1
  recorded_at: '2026-10-29T10:46:34.831Z'
semantic_text: A statement upload must go to the connector instance bound to the account.
logic_claims:
  - CLAIM-071AF3015FD52092
semantic_clauses:
  - A statement upload must go to the connector instance bound to the account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: b6501c69c1efb83569dd90229c85b7849ff299cadbb01f6bf09acf6b0e5096c8
semantic_inventory:
  - claim_key: CLAIM-071AF3015FD52092
    claim_text: A statement upload must go to the connector instance bound to the account
    role: normative
    status: modeled
    span:
      start: 0
      end: 73
id: REQ-ingest-statement-upload-bound-instance
type: req
---
A statement upload must go to the connector instance bound to the account.

## Context

D-79 makes statement uploads one of the owner's entry services, and D-49 has the core hand an upload to the right connector. D-109 decides how an upload is routed and booked. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> A statement upload goes to the connector instance bound to the account. A statement that covers several accounts, such as an IBKR Flex statement of linked accounts, books every mapped account bound to that instance and reports the broker accounts that aren't mapped (D-72). The core chooses the mode, opening for an account not yet seeded and transactions after that (D-95), and the owner may override it; an opening upload seeds the account for scheduled syncs too. The owner's UI learns through its owner service (D-79) which accounts have a connector that takes files, and shows the upload form only there. The core accepts files up to 64 MiB.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-109#1
