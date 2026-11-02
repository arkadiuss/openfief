---
title: Read access must never include any change of state.
status: closed
text_ref: design-decisions:D-137#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-137-4
  - lane:strict
semantic_text: Read access must never include any change of state.
logic_claims:
  - CLAIM-8A7C1BEDC6799147
semantic_clauses:
  - Read access must never include any change of state.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: baa6caa6875540692a377ca30ec8b08c04ca1699e421f7195b40d4531af5e72d
semantic_inventory:
  - claim_key: CLAIM-8A7C1BEDC6799147
    claim_text: Read access must never include any change of state
    role: normative
    status: modeled
    span:
      start: 0
      end: 50
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-137#4
  recorded_at: '2026-11-01T07:40:24.265Z'
id: REQ-identity-read-access-no-state-change
type: req
---
Read access must never include any change of state.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. D-6 makes the public API read-only, and data enters only through each connector's ingest path (D-7). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Read access is everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins. It never includes a secret, credential, passkey, session or token, a broker account number (D-121), or any change of state. Script tokens, view plugins and demo sessions hold it, and nothing else does.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-137#4
