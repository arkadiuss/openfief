---
title: An uploaded statement must be refused before a connector reads it when the caller is not the owner or the file is empty, and the connector must get only the cleaned last element of its name and a sniffed media type.
status: open
text_ref: docs/DESIGN.md#D-109
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-109
  - lane:strict
semantic_text: An uploaded statement must be refused before a connector reads it when the caller is not the owner or the file is empty, and the connector must get only the cleaned last element of its name and a sniffed media type.
logic_claims:
  - CLAIM-76780D141DF12E32
semantic_clauses:
  - An uploaded statement must be refused before a connector reads it when the caller is not the owner or the file is empty, and the connector must get only the cleaned last element of its name and a sniffed media type.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6d42ca7a6a474549996332fe9dc60bc9bc062a80783da3040b61933500846435
semantic_inventory:
  - claim_key: CLAIM-76780D141DF12E32
    claim_text: An uploaded statement must be refused before a connector reads it when the caller is not the owner or the file is empty, and the connector must get only the cleaned last element of its name and a sniffed media type
    role: normative
    status: modeled
    span:
      start: 0
      end: 214
origin:
  kind: agent
  recorded_at: '2026-10-30T04:11:11.980Z'
id: REQ-ingest-statement-upload-checks
type: req
---
An uploaded statement must be refused before a connector reads it when the caller is not the owner or the file is empty, and the connector must get only the cleaned last element of its name and a sniffed media type.

## Context

Restores, as a refinement, details of the superseded REQ-ingest-statement-upload-limits that the firm D-109 requirements don't state: owner only (D-79), no empty file, and the file's name and type only as hints.

## Source

> A statement upload goes to the connector instance bound to the account. A statement that covers several accounts, such as an IBKR Flex statement of linked accounts, books every mapped account bound to that instance and reports the broker accounts that aren't mapped (D-72). The core chooses the mode, opening for an account not yet seeded and transactions after that (D-95), and the owner may override it; an opening upload seeds the account for scheduled syncs too. The owner's UI learns through its owner service (D-79) which accounts have a connector that takes files, and shows the upload form only there. The core accepts files up to 64 MiB.

docs/DESIGN.md#D-109
