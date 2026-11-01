---
title: Read access must be held by script tokens, view plugins and demo sessions, and by nothing else.
status: open
text_ref: design-decisions:D-137#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-137-5
  - lane:strict
semantic_text: Read access must be held by script tokens, view plugins and demo sessions, and by nothing else.
logic_claims:
  - CLAIM-A0DDC43A746A5222
semantic_clauses:
  - Read access must be held by script tokens, view plugins and demo sessions, and by nothing else.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 03e7f4bf37bd5e818903e67465d6c5cd5a781d56655d60e5f75d038abf02ec35
semantic_inventory:
  - claim_key: CLAIM-A0DDC43A746A5222
    claim_text: Read access must be held by script tokens, view plugins and demo sessions, and by nothing else
    role: normative
    status: modeled
    span:
      start: 0
      end: 94
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-137#5
  recorded_at: '2026-11-01T07:40:30.828Z'
id: REQ-identity-read-access-holders
type: req
---
Read access must be held by script tokens, view plugins and demo sessions, and by nothing else.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. Script tokens are D-27 and D-67's, view plugins' own tokens are accepted by D-83, and demo sessions are D-68's read-scoped sessions. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Read access is everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins. It never includes a secret, credential, passkey, session or token, a broker account number (D-121), or any change of state. Script tokens, view plugins and demo sessions hold it, and nothing else does.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-137#5
