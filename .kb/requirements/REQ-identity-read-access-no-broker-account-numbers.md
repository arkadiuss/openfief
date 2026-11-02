---
title: Read access must never include a broker account number.
status: closed
text_ref: design-decisions:D-137#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-137-3
  - lane:strict
semantic_text: Read access must never include a broker account number.
logic_claims:
  - CLAIM-F68E743F700ECB9B
semantic_clauses:
  - Read access must never include a broker account number.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 19987c85253d65900b3106aa7a136059089bdcd512657e28045a9ec5a6106103
semantic_inventory:
  - claim_key: CLAIM-F68E743F700ECB9B
    claim_text: Read access must never include a broker account number
    role: normative
    status: modeled
    span:
      start: 0
      end: 54
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-137#3
  recorded_at: '2026-11-01T07:40:17.998Z'
id: REQ-identity-read-access-no-broker-account-numbers
type: req
---
Read access must never include a broker account number.

## Context

The author decided it on 2026-11-01, after reviewing accounts, bond payouts and read access. D-121 already keeps broker account numbers out of the read API (REQ-views-read-api-no-broker-account-numbers) and lists them only through the owner's account service. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Read access is everything the read API serves: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, transfers, attention items and the installed plugins. It never includes a secret, credential, passkey, session or token, a broker account number (D-121), or any change of state. Script tokens, view plugins and demo sessions hold it, and nothing else does.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-137#3
