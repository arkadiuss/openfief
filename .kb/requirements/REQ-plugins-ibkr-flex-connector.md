---
title: The IBKR Flex plugin must answer Sync by pulling its Flex Query from the Flex Web Service with the token and query ID from its mounted Secret, read on every sync and never logged or put in an error, and ImportFile by mapping an uploaded Flex statement, both with the ibkrflex mapper in the mode the call names, reaching only ndcdyn.interactivebrokers.com.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The IBKR Flex plugin must answer Sync by pulling its Flex Query from the Flex Web Service with the token and query ID from its mounted Secret, read on every sync and never logged or put in an error, and ImportFile by mapping an uploaded Flex statement, both with the ibkrflex mapper in the mode the call names, reaching only ndcdyn.interactivebrokers.com.
logic_claims:
  - CLAIM-8BF9471BDBE8A3CE
semantic_clauses:
  - The IBKR Flex plugin must answer Sync by pulling its Flex Query from the Flex Web Service with the token and query ID from its mounted Secret, read on every sync and never logged or put in an error, and ImportFile by mapping an uploaded Flex statement, both with the ibkrflex mapper in the mode the call names, reaching only ndcdyn.interactivebrokers.com.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 006f19543abf4ec0dc0955e5c37a8c142a9aad0e32182a665b9356d9c91fe937
semantic_inventory:
  - claim_key: CLAIM-8BF9471BDBE8A3CE
    claim_text: The IBKR Flex plugin must answer Sync by pulling its Flex Query from the Flex Web Service with the token and query ID from its mounted Secret, read on every sync and never logged or put in an error, and ImportFile by mapping an uploaded Flex statement, both with the ibkrflex mapper in the mode the call names, reaching only ndcdyn.interactivebrokers.com
    role: normative
    status: modeled
    span:
      start: 0
      end: 354
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:31.892Z'
id: REQ-plugins-ibkr-flex-connector
type: req
---
The IBKR Flex plugin must answer Sync by pulling its Flex Query from the Flex Web Service with the token and query ID from its mounted Secret, read on every sync and never logged or put in an error, and ImportFile by mapping an uploaded Flex statement, both with the ibkrflex mapper in the mode the call names, reaching only ndcdyn.interactivebrokers.com.

## Context

D-9 lists IBKR among M0's connectors and D-49 makes connectors plugins the core calls to sync and to import files, owning the mode of each import batch (D-23). The Flex token is a read-only credential that still exposes the full account history (DESIGN 3.5), so it comes from a Secret the owner creates (D-29), is read per sync so rotation needs no restart, and never appears in logs or errors. Refines REQ-AUTO-FD618A602F7F9106 and REQ-ingest-core-calls-connectors. Modeled by Claude Code for the M0 plugin binaries task (agent A: SDK verifier, operator key publishing, first-party plugins); the author has not reviewed it yet.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md D-49
