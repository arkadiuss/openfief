---
title: Worker concurrency, the daily end-of-day time with its time zone and the jurisdiction pack the workers value with must be configured for the revaluation workers, and invalid settings must be rejected.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: Worker concurrency, the daily end-of-day time with its time zone and the jurisdiction pack the workers value with must be configured for the revaluation workers, and invalid settings must be rejected.
logic_claims:
  - CLAIM-6F3106B1CB2F2890
semantic_clauses:
  - Worker concurrency, the daily end-of-day time with its time zone and the jurisdiction pack the workers value with must be configured for the revaluation workers, and invalid settings must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 96361a105020813aec886c16dd9d2824ad737901d51c3b304eb89d9337bc45aa
semantic_inventory:
  - claim_key: CLAIM-6F3106B1CB2F2890
    claim_text: Worker concurrency, the daily end-of-day time with its time zone and the jurisdiction pack the workers value with must be configured for the revaluation workers, and invalid settings must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 199
origin:
  kind: agent
  recorded_at: '2026-10-29T21:49:32.845Z'
id: REQ-valuation-revaluation-config
type: req
---
Worker concurrency, the daily end-of-day time with its time zone and the jurisdiction pack the workers value with must be configured for the revaluation workers, and invalid settings must be rejected.

## Context

Supersedes REQ-valuation-settings: D-77 removed the bond basis from what D-36 makes configurable. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-valuation-settings: D-77 removed the bond basis from the settings. The superseded requirement is closed.

## Source

> **[D-36]** Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time. [...] Configurable: worker concurrency and the end-of-day schedule. Each stored valuation holds all three values of D-77. What triggers a revaluation is not configurable.

docs/DESIGN.md#D-36
