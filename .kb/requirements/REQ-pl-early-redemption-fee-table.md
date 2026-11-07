---
title: A series' early-redemption fee must come from the reviewed table of the issue letters by bond type and sale month, and the table must give no fee for a month it was not reviewed for.
status: closed
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: A series' early-redemption fee must come from the reviewed table of the issue letters by bond type and sale month, and the table must give no fee for a month it was not reviewed for.
logic_claims:
  - CLAIM-6E5573677D94B2D3
semantic_clauses:
  - A series' early-redemption fee must come from the reviewed table of the issue letters by bond type and sale month, and the table must give no fee for a month it was not reviewed for.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5aba5b55fd3a7efc3ccb853cf64bd40c338c654904737054cb24892d57c6b169
semantic_inventory:
  - claim_key: CLAIM-6E5573677D94B2D3
    claim_text: A series' early-redemption fee must come from the reviewed table of the issue letters by bond type and sale month, and the table must give no fee for a month it was not reviewed for
    role: normative
    status: modeled
    span:
      start: 0
      end: 181
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:32.297Z'
id: REQ-pl-early-redemption-fee-table
type: req
---
A series' early-redemption fee must come from the reviewed table of the issue letters by bond type and sale month, and the table must give no fee for a month it was not reviewed for.

## Context

Author decision D-97, given after R-27 found obligacjeskarbowe.pl carries no grant of use while the ministry publishes the spreadsheet and the issue letters itself. It refines the firm D-97 requirements. The rows were read from every issue letter (R-28); a series after the reviewed month needs the owner's fee. Modeled by Claude Code for the D-97 bond terms work; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97
