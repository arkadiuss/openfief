---
title: The core must hand a statement file the owner uploads for an account to the one connector instance bound to that account through ImportFile, and import its delivery as an upload batch confined to that account.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The core must hand a statement file the owner uploads for an account to the one connector instance bound to that account through ImportFile, and import its delivery as an upload batch confined to that account.
logic_claims:
  - CLAIM-C0268BA484709DA0
semantic_clauses:
  - The core must hand a statement file the owner uploads for an account to the one connector instance bound to that account through ImportFile, and import its delivery as an upload batch confined to that account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 33c0eda7666c4280f2a62edf8a537a6650f3f308e42d52b59b695589d984c015
semantic_inventory:
  - claim_key: CLAIM-C0268BA484709DA0
    claim_text: The core must hand a statement file the owner uploads for an account to the one connector instance bound to that account through ImportFile, and import its delivery as an upload batch confined to that account
    role: normative
    status: modeled
    span:
      start: 0
      end: 208
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:22.124Z'
id: REQ-ingest-statement-upload-handoff
type: req
---
The core must hand a statement file the owner uploads for an account to the one connector instance bound to that account through ImportFile, and import its delivery as an upload batch confined to that account.

## Context

D-49: file uploads in the built-in UI go to the core, which hands them to the right connector; D-7 makes upload a connector ingest path. The right connector is read as the one instance bound to the account, as for bond lots; confining the delivery to that account and letting the owner pick the mode are this task's narrow readings, listed for the author. Modeled by Claude Code for the statement upload task; the author has not reviewed it yet.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md#D-49
