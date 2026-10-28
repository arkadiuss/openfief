---
title: A connector's ingest path, from a statement file or a pull, must map the statement with that institution's connector and import the resulting transactions and balance assertions into the ledger as one import batch, and must import nothing when the institution has no connector, the statement cannot be read or the pull fails.
status: open
text_ref: docs/DESIGN.md#D-7
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-7
  - lane:strict
semantic_text: A connector's ingest path, from a statement file or a pull, must map the statement with that institution's connector and import the resulting transactions and balance assertions into the ledger as one import batch, and must import nothing when the institution has no connector, the statement cannot be read or the pull fails.
logic_claims:
  - CLAIM-4007313BD110B132
semantic_clauses:
  - A connector's ingest path, from a statement file or a pull, must map the statement with that institution's connector and import the resulting transactions and balance assertions into the ledger as one import batch, and must import nothing when the institution has no connector, the statement cannot be read or the pull fails.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 01ee87b9f2e3694fd515bff3edb169b7010df0fee26225aebeda6e6c91ff72e1
semantic_inventory:
  - claim_key: CLAIM-4007313BD110B132
    claim_text: A connector's ingest path, from a statement file or a pull, must map the statement with that institution's connector and import the resulting transactions and balance assertions into the ledger as one import batch, and must import nothing when the institution has no connector, the statement cannot be read or the pull fails
    role: normative
    status: modeled
    span:
      start: 0
      end: 324
origin:
  kind: agent
  recorded_at: '2026-10-28T06:37:29.769Z'
id: REQ-ingest-app-ingest-path
type: req
---
A connector's ingest path, from a statement file or a pull, must map the statement with that institution's connector and import the resulting transactions and balance assertions into the ledger as one import batch, and must import nothing when the institution has no connector, the statement cannot be read or the pull fails.

## Context

D-7 lets data enter only through each connector's ingest path (upload, CLI, watched folder), never through the public API; the daily pull is the IBKR connector's other path. The ledger application is consumed through a port declared here, so the ingest context does not depend on the ledger's storage. Refines REQ-AUTO-41A1F4D687DAA771 and REQ-AUTO-D9FBE33256704D04. Modeled by Claude Code for the M0 IBKR Flex connector task; the author has not reviewed it yet.

## Source

> Data enters only through each connector's ingest path (upload, CLI, watched folder), never through the public API.

docs/DESIGN.md D-7
