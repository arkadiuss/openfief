---
title: Entries a connector names by broker account must be booked into the OpenFief account the broker account is mapped to, with ':' and that account ID appended to each transaction's key source, and nothing of a broker account that is not mapped may be booked.
status: open
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
semantic_text: Entries a connector names by broker account must be booked into the OpenFief account the broker account is mapped to, with ':' and that account ID appended to each transaction's key source, and nothing of a broker account that is not mapped may be booked.
logic_claims:
  - CLAIM-EAE5FFF3BA0B86BE
semantic_clauses:
  - Entries a connector names by broker account must be booked into the OpenFief account the broker account is mapped to, with ':' and that account ID appended to each transaction's key source, and nothing of a broker account that is not mapped may be booked.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bcdf92b0773a1b754553616e1fa88e719572958573ff7d0bf412befacee38c20
semantic_inventory:
  - claim_key: CLAIM-EAE5FFF3BA0B86BE
    claim_text: Entries a connector names by broker account must be booked into the OpenFief account the broker account is mapped to, with ':' and that account ID appended to each transaction's key source, and nothing of a broker account that is not mapped may be booked
    role: normative
    status: modeled
    span:
      start: 0
      end: 254
origin:
  kind: agent
  recorded_at: '2026-10-29T11:36:53.755Z'
id: REQ-ingest-broker-account-resolve
type: req
---
Entries a connector names by broker account must be booked into the OpenFief account the broker account is mapped to, with ':' and that account ID appended to each transaction's key source, and nothing of a broker account that is not mapped may be booked.

## Context

D-72: transactions of an unmapped broker account are not booked. Appending the account to the key source keeps keys per account as connectors key them (ledger.TransactionKey), so an IBKR account mapped in the UI gets the same keys the plugin's old account map gave. Refines REQ-ingest-unmapped-broker-account. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-72]** The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

docs/DESIGN.md#D-72
