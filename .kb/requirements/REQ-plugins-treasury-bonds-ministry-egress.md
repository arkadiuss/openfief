---
title: The treasury-bonds manifest must ask for egress to www.gov.pl and api.dane.gov.pl, with https-only base URL fields for them, and never for obligacjeskarbowe.pl.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: The treasury-bonds manifest must ask for egress to www.gov.pl and api.dane.gov.pl, with https-only base URL fields for them, and never for obligacjeskarbowe.pl.
logic_claims:
  - CLAIM-BA4A5E8E6AD9D8F4
semantic_clauses:
  - The treasury-bonds manifest must ask for egress to www.gov.pl and api.dane.gov.pl, with https-only base URL fields for them, and never for obligacjeskarbowe.pl.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a6f3c5c5a2b3c47a173df2f586b354aea44b4393747215b3f77e16ae690b2cca
semantic_inventory:
  - claim_key: CLAIM-BA4A5E8E6AD9D8F4
    claim_text: The treasury-bonds manifest must ask for egress to www.gov.pl and api.dane.gov.pl, with https-only base URL fields for them, and never for obligacjeskarbowe.pl
    role: normative
    status: modeled
    span:
      start: 0
      end: 159
origin:
  kind: agent
  recorded_at: '2026-11-07T08:07:13.657Z'
id: REQ-plugins-treasury-bonds-ministry-egress
type: req
---
The treasury-bonds manifest must ask for egress to www.gov.pl and api.dane.gov.pl, with https-only base URL fields for them, and never for obligacjeskarbowe.pl.

## Context

Author decision D-97, given after R-27 found obligacjeskarbowe.pl carries no grant of use while the ministry publishes the spreadsheet and the issue letters itself. It refines the firm D-97 requirements. Modeled by Claude Code for the D-97 bond terms work; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97
