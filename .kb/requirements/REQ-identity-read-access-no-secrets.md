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
  recorded_at: '2026-11-01T07:40:11.886Z'
id: REQ-identity-read-access-no-secrets
type: req
---
Read access must never include a secret, credential, passkey, session or token.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. The security section's threats include leaked read-only credentials, which still expose the full account history; what they expose stops at the record itself. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Read access is everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins. It never includes a secret, credential, passkey, session or token, a broker account number (D-121), or any change of state. Script tokens, view plugins and demo sessions hold it, and nothing else does.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-137#2
