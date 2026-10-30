---
title: The core must read its database connection string from OPENFIEF_DSN and, from flags with documented defaults, its browser port's and plugin port's listen addresses, reporting currency, the jurisdiction pack's by default, time zone, end-of-day time, revaluation concurrency and shutdown timeout, and refuse to start on a missing or invalid value.
status: open
text_ref: docs/DESIGN.md#D-62
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-62
  - lane:strict
semantic_text: The core must read its database connection string from OPENFIEF_DSN and, from flags with documented defaults, its browser port's and plugin port's listen addresses, reporting currency, the jurisdiction pack's by default, time zone, end-of-day time, revaluation concurrency and shutdown timeout, and refuse to start on a missing or invalid value.
logic_claims:
  - CLAIM-78D6A6E77268C2C9
semantic_clauses:
  - The core must read its database connection string from OPENFIEF_DSN and, from flags with documented defaults, its browser port's and plugin port's listen addresses, reporting currency, the jurisdiction pack's by default, time zone, end-of-day time, revaluation concurrency and shutdown timeout, and refuse to start on a missing or invalid value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ce9d892f1d0478eadea5f0fea011e4c26e9bc10111af23cc4a622b5ac606e49d
semantic_inventory:
  - claim_key: CLAIM-78D6A6E77268C2C9
    claim_text: The core must read its database connection string from OPENFIEF_DSN and, from flags with documented defaults, its browser port's and plugin port's listen addresses, reporting currency, the jurisdiction pack's by default, time zone, end-of-day time, revaluation concurrency and shutdown timeout, and refuse to start on a missing or invalid value
    role: normative
    status: modeled
    span:
      start: 0
      end: 344
origin:
  kind: agent
  recorded_at: '2026-10-29T21:49:41.225Z'
id: REQ-core-serve-flags
type: req
---
The core must read its database connection string from OPENFIEF_DSN and, from flags with documented defaults, its browser port's and plugin port's listen addresses, reporting currency, the jurisdiction pack's by default, time zone, end-of-day time, revaluation concurrency and shutdown timeout, and refuse to start on a missing or invalid value.

## Context

Supersedes REQ-core-config: the bond basis flag is gone with D-77, and the reporting currency defaults to the pack's (D-62). Modeled by Claude Code for the values work package; the author has not reviewed it yet. D-115 gives the core a browser port (--listen, :8080) and a plugin port (--plugin-listen, :8081).

## Why this changed

It replaces REQ-core-config: The bond basis flag is gone with D-77. The superseded requirement is closed.

## Source

> **[D-62]** Totals, charts and the model comparison are shown in the reporting currency: a setting whose default comes from the jurisdiction pack, PLN for the Polish pack. Apart from the Polish pack's tax rules, no code assumes PLN.

docs/DESIGN.md#D-62
