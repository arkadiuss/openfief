---
title: The owner's UI must show the statement upload form only on accounts whose bound connector takes files, learning which through its owner service.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-109-5
  - lane:strict
text_ref: design-decisions:D-109#5
origin:
  kind: import
  ref: design-decisions:D-109#5
  recorded_at: '2026-10-29T10:49:52.844Z'
semantic_text: The owner's UI must show the statement upload form only on accounts whose bound connector takes files, learning which through its owner service.
logic_claims:
  - CLAIM-18D09856721263FB
semantic_clauses:
  - The owner's UI must show the statement upload form only on accounts whose bound connector takes files, learning which through its owner service.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: b9f98625d0deda269e94fa00e1c3c6f1d64097aac386ea7ad13a4cd5a144a6f5
semantic_inventory:
  - claim_key: CLAIM-18D09856721263FB
    claim_text: The owner's UI must show the statement upload form only on accounts whose bound connector takes files, learning which through its owner service
    role: normative
    status: modeled
    span:
      start: 0
      end: 143
id: REQ-ui-statement-upload-file-connectors-only
type: req
---
The owner's UI must show the statement upload form only on accounts whose bound connector takes files, learning which through its owner service.

## Why this changed

The author's decision D-109 has the owner's UI learn through its owner service (D-79) which accounts have a connector that takes files, and show the upload form only there. REQ-ui-statement-upload offered the form on every account page, so this requirement replaces it.

## Context

D-79 makes statement uploads one of the owner's entry services, called only by a view plugin acting for an owner session. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> A statement upload goes to the connector instance bound to the account. A statement that covers several accounts, such as an IBKR Flex statement of linked accounts, books every mapped account bound to that instance and reports the broker accounts that aren't mapped (D-72). The core chooses the mode, opening for an account not yet seeded and transactions after that (D-95), and the owner may override it; an opening upload seeds the account for scheduled syncs too. The owner's UI learns through its owner service (D-79) which accounts have a connector that takes files, and shows the upload form only there. The core accepts files up to 64 MiB.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-109#5
