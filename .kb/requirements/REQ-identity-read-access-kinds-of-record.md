---
title: 'Read access must be everything the read API serves, named by kind of record: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, balance assertion checks, transfers, attention items, yearly taxes, the settings with each account''s tax wrapper, and the installed plugins.'
status: open
text_ref: design-decisions:D-137#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-137-1
  - lane:strict
semantic_text: 'Read access must be everything the read API serves, named by kind of record: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, balance assertion checks, transfers, attention items, yearly taxes, the settings with each account''s tax wrapper, and the installed plugins.'
logic_claims:
  - CLAIM-4006991836C61A68
semantic_clauses:
  - 'Read access must be everything the read API serves, named by kind of record: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, balance assertion checks, transfers, attention items, yearly taxes, the settings with each account''s tax wrapper, and the installed plugins.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cad0330ab84f9b752d7032b1107e4dae48a821c49d24564fb61a450781c24d2a
semantic_inventory:
  - claim_key: CLAIM-4006991836C61A68
    claim_text: 'Read access must be everything the read API serves, named by kind of record: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, balance assertion checks, transfers, attention items, yearly taxes, the settings with each account''s tax wrapper, and the installed plugins'
    role: normative
    status: modeled
    span:
      start: 0
      end: 386
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-137#1
  recorded_at: '2026-11-02T00:53:30.812Z'
id: REQ-identity-read-access-kinds-of-record
type: req
---
Read access must be everything the read API serves, named by kind of record: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, balance assertion checks, transfers, attention items, yearly taxes, the settings with each account's tax wrapper, and the installed plugins.

## Why this changed

On 2026-11-01 the author rewrote D-137 (reason not stated): read access is named by kind of record, adds balance assertion checks, yearly taxes and the settings with each account's tax wrapper, counts the owner among its holders, and limits each holder to the accounts it may view (D-163). This claim now lists the records by kind and adds those three, so this requirement replaces REQ-identity-read-access-scope, which quoted the earlier text, and it is closed.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. The read API (D-50) already serves every kind listed, including the settings, yearly taxes and balance assertion checks. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Read access is everything the read API serves, named by kind of record: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, balance assertion checks, transfers, attention items, yearly taxes, the settings with each account's tax wrapper, and the installed plugins. It never includes a secret, credential, passkey, session or token, a broker account number (D-121), or any change of state. The owner, script tokens, view plugins and demo sessions hold it, and nothing else does; each reads only the accounts it may view (D-163).

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-137#1
