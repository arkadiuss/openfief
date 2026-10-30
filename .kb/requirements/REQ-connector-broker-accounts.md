---
title: The connector contract must let a transaction and a balance assertion name a broker account instead of an OpenFief account, never both, and let a delivery list every broker account it saw, as additions only, and the core must refuse a delivery naming both or a malformed broker account.
status: open
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
semantic_text: The connector contract must let a transaction and a balance assertion name a broker account instead of an OpenFief account, never both, and let a delivery list every broker account it saw, as additions only, and the core must refuse a delivery naming both or a malformed broker account.
logic_claims:
  - CLAIM-9251EB4C9F549979
semantic_clauses:
  - The connector contract must let a transaction and a balance assertion name a broker account instead of an OpenFief account, never both, and let a delivery list every broker account it saw, as additions only, and the core must refuse a delivery naming both or a malformed broker account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 11280ebc99abc3166e412a70dc1681663b58b01474899708c333dd1da5ae62e6
semantic_inventory:
  - claim_key: CLAIM-9251EB4C9F549979
    claim_text: The connector contract must let a transaction and a balance assertion name a broker account instead of an OpenFief account, never both, and let a delivery list every broker account it saw, as additions only, and the core must refuse a delivery naming both or a malformed broker account
    role: normative
    status: modeled
    span:
      start: 0
      end: 285
origin:
  kind: agent
  recorded_at: '2026-10-29T11:37:08.421Z'
id: REQ-connector-broker-accounts
type: req
---
The connector contract must let a transaction and a balance assertion name a broker account instead of an OpenFief account, never both, and let a delivery list every broker account it saw, as additions only, and the core must refuse a delivery naming both or a malformed broker account.

## Context

D-72 needs the connector to report broker accounts and the core to map them, so the owner never types an account number into a plugin's config. buf breaking passes: only fields and a message are added. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-72]** The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

docs/DESIGN.md#D-72
