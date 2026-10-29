---
title: The core must serve StatementService to view plugins only as a client stream, header first and at most 1 MiB per message, and answer the batch's counts with at most 100 ingest problems or each failure's own code.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The core must serve StatementService to view plugins only as a client stream, header first and at most 1 MiB per message, and answer the batch's counts with at most 100 ingest problems or each failure's own code.
logic_claims:
  - CLAIM-E93BA51E161C34FE
semantic_clauses:
  - The core must serve StatementService to view plugins only as a client stream, header first and at most 1 MiB per message, and answer the batch's counts with at most 100 ingest problems or each failure's own code.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cd666776d68ef7fcc7cce2bbe2c33a21140aaaafe33f094e7da284a87aaabdc3
semantic_inventory:
  - claim_key: CLAIM-E93BA51E161C34FE
    claim_text: The core must serve StatementService to view plugins only as a client stream, header first and at most 1 MiB per message, and answer the batch's counts with at most 100 ingest problems or each failure's own code
    role: normative
    status: modeled
    span:
      start: 0
      end: 211
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:43.110Z'
id: REQ-ingest-statement-service
type: req
---
The core must serve StatementService to view plugins only as a client stream, header first and at most 1 MiB per message, and answer the batch's counts with at most 100 ingest problems or each failure's own code.

## Context

D-49 routes uploads from the UI through the core; the service follows the BondLotService (view plugins only, owner session in the request). A client stream keeps a 64 MiB file out of memory end to end; this choice and the problem cap are this task's, listed for the author. Modeled by Claude Code for the statement upload task; the author has not reviewed it yet.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md#D-49
