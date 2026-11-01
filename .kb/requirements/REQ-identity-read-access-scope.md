---
title: 'Read access must be everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins.'
status: open
text_ref: design-decisions:D-137#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-137-1
  - lane:strict
semantic_text: 'Read access must be everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins.'
logic_claims:
  - CLAIM-A64525E68682D0DE
semantic_clauses:
  - 'Read access must be everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bb0f142572e8a6aad1c83ecba8bbe07b0da5573d5fe2ecee222d6f66c1796689
semantic_inventory:
  - claim_key: CLAIM-A64525E68682D0DE
    claim_text: 'Read access must be everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins'
    role: normative
    status: modeled
    span:
      start: 0
      end: 274
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-137#1
  recorded_at: '2026-11-01T07:40:06.087Z'
id: REQ-identity-read-access-scope
type: req
---
Read access must be everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. D-66 gives each session a scope, read or owner, and D-83 says who the read API accepts. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Read access is everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins. It never includes a secret, credential, passkey, session or token, a broker account number (D-121), or any change of state. Script tokens, view plugins and demo sessions hold it, and nothing else does.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-137#1
