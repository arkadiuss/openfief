---
title: A statement upload must be refused unless the caller may change every account bound to the account's connector instance, since a statement of several accounts may book any of them.
status: open
text_ref: docs/DESIGN.md#D-163
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-163
  - lane:strict
semantic_text: A statement upload must be refused unless the caller may change every account bound to the account's connector instance, since a statement of several accounts may book any of them.
logic_claims:
  - CLAIM-54FFAB7294A2F31F
semantic_clauses:
  - A statement upload must be refused unless the caller may change every account bound to the account's connector instance, since a statement of several accounts may book any of them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6891aafd8951451817c7a2ec2b159d1070807c340a0ea181db31df07906ae570
semantic_inventory:
  - claim_key: CLAIM-54FFAB7294A2F31F
    claim_text: A statement upload must be refused unless the caller may change every account bound to the account's connector instance, since a statement of several accounts may book any of them
    role: exception
    status: modeled
    span:
      start: 0
      end: 179
origin:
  kind: agent
  recorded_at: '2026-11-02T13:17:43.357Z'
id: REQ-ingest-multi-account-upload-changes-every-bound
type: req
---
A statement upload must be refused unless the caller may change every account bound to the account's connector instance, since a statement of several accounts may book any of them.

## Context

D-140 has a statement of several accounts book every mapped account bound to the instance, and D-163 has every owner service ask the authorizer whether the caller may act on each account it changes. The core asks before it calls the connector, so it can't know which accounts the statement books and asks for all those bound to the instance; this is the implementation's reading, not stated by the author. Until SpiceDB answers (D-164) the owner may change every account, so nothing is refused today.

## Source

> **[D-140]** An upload of a statement covering several accounts runs in the mode chosen for the account it was uploaded for, which the owner can override: an opening upload seeds only the accounts not seeded yet, and a transactions upload books every mapped account.

docs/DESIGN.md#D-140

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, ...

docs/DESIGN.md#D-163
