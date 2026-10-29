---
title: The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, the jurisdiction pack's by default, time zone, end-of-day time, revaluation concurrency and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value.
status: open
text_ref: docs/DESIGN.md#D-62
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-62
  - lane:strict
semantic_text: The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, the jurisdiction pack's by default, time zone, end-of-day time, revaluation concurrency and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value.
logic_claims:
  - CLAIM-CC16AA39A7C1638B
semantic_clauses:
  - The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, the jurisdiction pack's by default, time zone, end-of-day time, revaluation concurrency and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6c45b4088b6bc0a3a8c3994347405b1cad3d9e2395d28fab195c681fc7fd99b9
semantic_inventory:
  - claim_key: CLAIM-CC16AA39A7C1638B
    claim_text: The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, the jurisdiction pack's by default, time zone, end-of-day time, revaluation concurrency and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value
    role: normative
    status: modeled
    span:
      start: 0
      end: 307
origin:
  kind: agent
  recorded_at: '2026-10-29T21:49:41.225Z'
id: REQ-core-serve-flags
type: req
---
The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, the jurisdiction pack's by default, time zone, end-of-day time, revaluation concurrency and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value.

## Context

Supersedes REQ-core-config: the bond basis flag is gone with D-77, and the reporting currency defaults to the pack's (D-62). Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-62]** Totals, charts and the model comparison are shown in the reporting currency: a setting whose default comes from the jurisdiction pack, PLN for the Polish pack. Apart from the Polish pack's tax rules, no code assumes PLN.

docs/DESIGN.md#D-62
