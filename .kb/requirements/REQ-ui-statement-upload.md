---
title: The built-in UI must offer the owner, and only the owner, an Upload statement form on every account page, stream the posted file to the core without storing it, and show the import's counts and problems or the core's reason.
status: closed
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The built-in UI must offer the owner, and only the owner, an Upload statement form on every account page, stream the posted file to the core without storing it, and show the import's counts and problems or the core's reason.
logic_claims:
  - CLAIM-D2190356B491ED64
semantic_clauses:
  - The built-in UI must offer the owner, and only the owner, an Upload statement form on every account page, stream the posted file to the core without storing it, and show the import's counts and problems or the core's reason.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 448f579c05dfba772f543b5310fc5ed1c8c8256d95e2d2ef52ccaa3d37f0e7b6
semantic_inventory:
  - claim_key: CLAIM-D2190356B491ED64
    claim_text: The built-in UI must offer the owner, and only the owner, an Upload statement form on every account page, stream the posted file to the core without storing it, and show the import's counts and problems or the core's reason
    role: normative
    status: modeled
    span:
      start: 0
      end: 223
origin:
  kind: agent
  recorded_at: '2026-10-29T05:39:15.846Z'
id: REQ-ui-statement-upload
type: req
---
The built-in UI must offer the owner, and only the owner, an Upload statement form on every account page, stream the posted file to the core without storing it, and show the import's counts and problems or the core's reason.

## Context

D-49: file uploads in the built-in UI go to the core. The UI does not know which connector is bound to an account, so it shows the form on every account page and the connector's refusal when it takes no files: this task's narrow reading. Modeled by Claude Code for the statement upload task; the author has not reviewed it yet.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md#D-49
