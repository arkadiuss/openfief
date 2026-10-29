---
title: The core must take a bond lot only from the owner's owner-scoped session, hand it to the one connector instance bound to its account as a new transactions batch, and import what that instance delivers into that account only, recording the delivered series terms.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The core must take a bond lot only from the owner's owner-scoped session, hand it to the one connector instance bound to its account as a new transactions batch, and import what that instance delivers into that account only, recording the delivered series terms.
logic_claims:
  - CLAIM-6AAC65AA02FC492F
semantic_clauses:
  - The core must take a bond lot only from the owner's owner-scoped session, hand it to the one connector instance bound to its account as a new transactions batch, and import what that instance delivers into that account only, recording the delivered series terms.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 292d8643c6d40a357f655bd398e61e4f09634cc42a09e338db290d8027e8b873
semantic_inventory:
  - claim_key: CLAIM-6AAC65AA02FC492F
    claim_text: The core must take a bond lot only from the owner's owner-scoped session, hand it to the one connector instance bound to its account as a new transactions batch, and import what that instance delivers into that account only, recording the delivered series terms
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:27.171Z'
id: REQ-ingest-bond-lot-handoff
type: req
---
The core must take a bond lot only from the owner's owner-scoped session, hand it to the one connector instance bound to its account as a new transactions batch, and import what that instance delivers into that account only, recording the delivered series terms.

## Context

D-49: manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector. The right connector is read as the one instance bound to the account; zero or several is a refusal. The owner-only session check copies the model portfolio's pattern, still awaiting the author. Modeled by Claude Code for the manual bond lot task; the author has not reviewed it yet.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md#D-49
