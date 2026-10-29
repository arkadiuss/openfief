---
title: A key, sleeve ID or account ID made from a name the owner typed must be the name in lower case, with Polish letters spelled in ASCII and every run of other characters one hyphen, none at either end, at most 64 characters, and making one from a name without a letter or digit must fail.
status: open
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
semantic_text: A key, sleeve ID or account ID made from a name the owner typed must be the name in lower case, with Polish letters spelled in ASCII and every run of other characters one hyphen, none at either end, at most 64 characters, and making one from a name without a letter or digit must fail.
logic_claims:
  - CLAIM-5367D2D4B21A9AF8
semantic_clauses:
  - A key, sleeve ID or account ID made from a name the owner typed must be the name in lower case, with Polish letters spelled in ASCII and every run of other characters one hyphen, none at either end, at most 64 characters, and making one from a name without a letter or digit must fail.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0a748b9c745d5204a463ced9b856f8bfefe397a8559667184a2bf428d0d2652c
semantic_inventory:
  - claim_key: CLAIM-5367D2D4B21A9AF8
    claim_text: A key, sleeve ID or account ID made from a name the owner typed must be the name in lower case, with Polish letters spelled in ASCII and every run of other characters one hyphen, none at either end, at most 64 characters, and making one from a name without a letter or digit must fail
    role: normative
    status: modeled
    span:
      start: 0
      end: 284
origin:
  kind: agent
  recorded_at: '2026-10-29T11:36:37.026Z'
id: REQ-portfolio-key-from-name
type: req
---
A key, sleeve ID or account ID made from a name the owner typed must be the name in lower case, with Polish letters spelled in ASCII and every run of other characters one hyphen, none at either end, at most 64 characters, and making one from a name without a letter or digit must fail.

## Context

The owner names kinds, sleeves and accounts in the UI (D-71, D-72) and never types an ID; the key form is the one sleeve kind keys already have, so it is safe in URLs and transaction key sources. The exact rule is the implementation's reading; no decision states it. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-72]** The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

docs/DESIGN.md#D-72
