---
title: 'The owner must set an OKI account''s form when creating it: a bank account, a treasury bond account, a brokerage account or a fund register.'
status: open
text_ref: design-decisions:D-147#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-147-1
  - lane:strict
semantic_text: 'The owner must set an OKI account''s form when creating it: a bank account, a treasury bond account, a brokerage account or a fund register.'
logic_claims:
  - CLAIM-020723432DD86FE6
semantic_clauses:
  - 'The owner must set an OKI account''s form when creating it: a bank account, a treasury bond account, a brokerage account or a fund register.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2c8d0e539296be8a803b7d7c4dfd840ee399e2780577b5f3f8a78e82794b1308
semantic_inventory:
  - claim_key: CLAIM-020723432DD86FE6
    claim_text: 'The owner must set an OKI account''s form when creating it: a bank account, a treasury bond account, a brokerage account or a fund register'
    role: normative
    status: modeled
    span:
      start: 0
      end: 138
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-147#1
  recorded_at: '2026-11-01T12:27:10.205Z'
id: REQ-ingest-oki-account-form-set-at-creation
type: req
---
The owner must set an OKI account's form when creating it: a bank account, a treasury bond account, a brokerage account or a fund register.

## Context

The author recorded it on 2026-11-01, answering the implementation's OKI questions (#40 to #43, #46). The owner creates accounts in the built-in UI (D-72), and OKI is one of the Polish pack's tax wrappers (REQ-pl-tax-wrappers). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner sets an OKI account's form when creating it: a bank account, a treasury bond account, a brokerage account or a fund register. The form decides whether its PLN cash is a savings or an investment asset.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-147#1
