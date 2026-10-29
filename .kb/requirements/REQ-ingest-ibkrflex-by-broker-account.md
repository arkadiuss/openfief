---
title: The IBKR Flex connector must take no account config, name every statement's entries by its IBKR account with keys whose source names no account, and report every IBKR account it saw, so that the core's mapping gives the same entries and keys as a configured account map.
status: open
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
semantic_text: The IBKR Flex connector must take no account config, name every statement's entries by its IBKR account with keys whose source names no account, and report every IBKR account it saw, so that the core's mapping gives the same entries and keys as a configured account map.
logic_claims:
  - CLAIM-E472BF36C257D7BB
semantic_clauses:
  - The IBKR Flex connector must take no account config, name every statement's entries by its IBKR account with keys whose source names no account, and report every IBKR account it saw, so that the core's mapping gives the same entries and keys as a configured account map.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 02305ffb81eb369ba9f59753604feef2d252356e94fec74e995b1b32d70d71c4
semantic_inventory:
  - claim_key: CLAIM-E472BF36C257D7BB
    claim_text: The IBKR Flex connector must take no account config, name every statement's entries by its IBKR account with keys whose source names no account, and report every IBKR account it saw, so that the core's mapping gives the same entries and keys as a configured account map
    role: normative
    status: modeled
    span:
      start: 0
      end: 269
origin:
  kind: agent
  recorded_at: '2026-10-29T11:37:14.915Z'
id: REQ-ingest-ibkrflex-by-broker-account
type: req
---
The IBKR Flex connector must take no account config, name every statement's entries by its IBKR account with keys whose source names no account, and report every IBKR account it saw, so that the core's mapping gives the same entries and keys as a configured account map.

## Context

D-72: the owner never types a broker account number; ibkr-flex's config used to map Flex account IDs to OpenFief accounts. Fixtures use synthetic account IDs only. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-72]** The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

docs/DESIGN.md#D-72
