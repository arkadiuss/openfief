---
title: The core must record the version of the jurisdiction pack's rules at start-up and revalue every account from its first transaction in the same transaction when the version differs from the recorded one, and nothing when it is the same.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: The core must record the version of the jurisdiction pack's rules at start-up and revalue every account from its first transaction in the same transaction when the version differs from the recorded one, and nothing when it is the same.
logic_claims:
  - CLAIM-FB3835928BD6B481
semantic_clauses:
  - The core must record the version of the jurisdiction pack's rules at start-up and revalue every account from its first transaction in the same transaction when the version differs from the recorded one, and nothing when it is the same.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1906809cdee45448c4cd255c8c2238bc2838d115167a5d02c94bc6590ed3cfb5
semantic_inventory:
  - claim_key: CLAIM-FB3835928BD6B481
    claim_text: The core must record the version of the jurisdiction pack's rules at start-up and revalue every account from its first transaction in the same transaction when the version differs from the recorded one, and nothing when it is the same
    role: normative
    status: modeled
    span:
      start: 0
      end: 234
origin:
  kind: agent
  recorded_at: '2026-10-29T21:49:17.602Z'
id: REQ-valuation-rules-version-revalues
type: req
---
The core must record the version of the jurisdiction pack's rules at start-up and revalue every account from its first transaction in the same transaction when the version differs from the recorded one, and nothing when it is the same.

## Context

Supersedes REQ-valuation-basis-change: D-36 no longer revalues on a basis change since D-77 removed the basis. A code change to the pack's rules changes stored values, so its version stands in for the basis. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-valuation-basis-change: D-36 no longer revalues on a basis change. The superseded requirement is closed.

## Source

> **[D-36]** Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time. [...] Configurable: worker concurrency and the end-of-day schedule. Each stored valuation holds all three values of D-77. What triggers a revaluation is not configurable.

docs/DESIGN.md#D-36
