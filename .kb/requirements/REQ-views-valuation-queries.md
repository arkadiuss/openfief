---
title: The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and PLN value, with totals per date, and the whole portfolio's totals per date summed over every account.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and PLN value, with totals per date, and the whole portfolio's totals per date summed over every account.
logic_claims:
  - CLAIM-A166372FEDD3BD08
semantic_clauses:
  - The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and PLN value, with totals per date, and the whole portfolio's totals per date summed over every account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 372d10620db55efd8e2c034726c651865e58b1a9e2ddf5ce0d6328c6c38629ce
semantic_inventory:
  - claim_key: CLAIM-A166372FEDD3BD08
    claim_text: The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and PLN value, with totals per date, and the whole portfolio's totals per date summed over every account
    role: normative
    status: modeled
    span:
      start: 0
      end: 247
origin:
  kind: agent
  recorded_at: '2026-10-28T07:25:56.705Z'
id: REQ-views-valuation-queries
type: req
---
The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and PLN value, with totals per date, and the whole portfolio's totals per date summed over every account.

## Context

D-6 puts valuations in the public API and D-36 makes them stored rows, so the views read them through a consumer-side Valuations port instead of computing them; the valuation context's store implements it when cmd wires it. Totals feed the value-over-time chart (D-30). Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it.

docs/DESIGN.md D-6

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time. [...] The read API reports, per account, the date it is valued through and whether a revaluation is pending or running, and the UI shows it.

docs/DESIGN.md D-36
