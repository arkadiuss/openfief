---
title: A statement row that a connector cannot or will not book must become an ingest problem, unsupported or skipped, naming the row and the reason without amounts or account numbers, and must not fail the rest of the import batch.
status: open
text_ref: docs/DESIGN.md#D-7
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-7
  - lane:strict
semantic_text: A statement row that a connector cannot or will not book must become an ingest problem, unsupported or skipped, naming the row and the reason without amounts or account numbers, and must not fail the rest of the import batch.
logic_claims:
  - CLAIM-C1B6E7573AB6C897
semantic_clauses:
  - A statement row that a connector cannot or will not book must become an ingest problem, unsupported or skipped, naming the row and the reason without amounts or account numbers, and must not fail the rest of the import batch.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 613a79fe729394a80db1b2492641660ebf39ed811017458495ea2b448ca55de0
semantic_inventory:
  - claim_key: CLAIM-C1B6E7573AB6C897
    claim_text: A statement row that a connector cannot or will not book must become an ingest problem, unsupported or skipped, naming the row and the reason without amounts or account numbers, and must not fail the rest of the import batch
    role: normative
    status: modeled
    span:
      start: 0
      end: 224
origin:
  kind: agent
  recorded_at: '2026-10-28T06:24:19.144Z'
id: REQ-ingest-problems-not-errors
type: req
---
A statement row that a connector cannot or will not book must become an ingest problem, unsupported or skipped, naming the row and the reason without amounts or account numbers, and must not fail the rest of the import batch.

## Context

Institutions report more than OpenFief books in M0 (options, futures, CFDs, interest, corrections). Failing the whole statement would block every import of a real account, while silently dropping rows would hide why a balance assertion fails; a recorded problem does neither. Problem references carry only the statement section, position and the institution's row ID, per GR-12. Modeled by Claude Code for the M0 IBKR Flex connector task; the author has not reviewed it yet. The Flex schema is IBKR's public Flex Query reference, not yet verified against a real account (OQ-22).

## Source

> Data enters only through each connector's ingest path (upload, CLI, watched folder), never through the public API.

docs/DESIGN.md D-7
