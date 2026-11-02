---
title: Read access must be held by the owner, script tokens, view plugins and demo sessions, and by nothing else.
status: open
text_ref: design-decisions:D-137#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-137-5
  - lane:strict
semantic_text: Read access must be held by the owner, script tokens, view plugins and demo sessions, and by nothing else.
logic_claims:
  - CLAIM-79931132DC178ED6
semantic_clauses:
  - Read access must be held by the owner, script tokens, view plugins and demo sessions, and by nothing else.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bee504bf5d464f87a618018107a81b1d2cc82c05f4130e9c90ac2b9dcd656139
semantic_inventory:
  - claim_key: CLAIM-79931132DC178ED6
    claim_text: Read access must be held by the owner, script tokens, view plugins and demo sessions, and by nothing else
    role: normative
    status: modeled
    span:
      start: 0
      end: 105
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-137#5
  recorded_at: '2026-11-02T00:53:55.835Z'
id: REQ-identity-read-access-holders-incl-owner
type: req
---
Read access must be held by the owner, script tokens, view plugins and demo sessions, and by nothing else.

## Why this changed

On 2026-11-01 the author rewrote D-137 (reason not stated): read access is named by kind of record, adds balance assertion checks, yearly taxes and the settings with each account's tax wrapper, counts the owner among its holders, and limits each holder to the accounts it may view (D-163). This claim adds the owner to the holders, so this requirement replaces REQ-identity-read-access-holders, which quoted the earlier text, and it is closed.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. D-83 has the read API accept sessions of either scope, so the owner's session already reads it; it refuses connector tokens. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Read access is everything the read API serves, named by kind of record: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, balance assertion checks, transfers, attention items, yearly taxes, the settings with each account's tax wrapper, and the installed plugins. It never includes a secret, credential, passkey, session or token, a broker account number (D-121), or any change of state. The owner, script tokens, view plugins and demo sessions hold it, and nothing else does; each reads only the accounts it may view (D-163).

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-137#5
