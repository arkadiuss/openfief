---
title: Read access must never include a secret, credential, passkey, session or token.
status: open
text_ref: design-decisions:D-137#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-137-2
  - lane:strict
semantic_text: Read access must never include a secret, credential, passkey, session or token.
logic_claims:
  - CLAIM-020306A178C5E6E9
semantic_clauses:
  - Read access must never include a secret, credential, passkey, session or token.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bf03d9dac866f92c4cdf7fccddac7efdacd29b2455d011e328037c1082e3e646
semantic_inventory:
  - claim_key: CLAIM-020306A178C5E6E9
    claim_text: Read access must never include a secret, credential, passkey, session or token
    role: normative
    status: modeled
    span:
      start: 0
      end: 78
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-137#2
  recorded_at: '2026-11-02T00:53:37.173Z'
id: REQ-identity-read-access-never-secrets
type: req
---
Read access must never include a secret, credential, passkey, session or token.

## Why this changed

On 2026-11-01 the author rewrote D-137 (reason not stated): read access is named by kind of record, adds balance assertion checks, yearly taxes and the settings with each account's tax wrapper, counts the owner among its holders, and limits each holder to the accounts it may view (D-163). This claim's wording is unchanged, so this requirement replaces REQ-identity-read-access-no-secrets, which quoted the earlier text, and it is closed.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. Script tokens, view plugins and demo visitors read through it, so nothing that authenticates may leak to them. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Read access is everything the read API serves, named by kind of record: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, balance assertion checks, transfers, attention items, yearly taxes, the settings with each account's tax wrapper, and the installed plugins. It never includes a secret, credential, passkey, session or token, a broker account number (D-121), or any change of state. The owner, script tokens, view plugins and demo sessions hold it, and nothing else does; each reads only the accounts it may view (D-163).

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-137#2
