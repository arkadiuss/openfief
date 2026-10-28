---
title: The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value.
logic_claims:
  - CLAIM-10A9FFBBD060315D
semantic_clauses:
  - The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 06a162e204c243e636a1574070aad58337e5b55e8a3e0d3ae68c6f71547aa283
semantic_inventory:
  - claim_key: CLAIM-10A9FFBBD060315D
    claim_text: The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value
    role: normative
    status: modeled
    span:
      start: 0
      end: 283
origin:
  kind: agent
  recorded_at: '2026-10-28T14:33:18.485Z'
id: REQ-core-config
type: req
---
The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value.

## Context

D-36 makes concurrency, the end-of-day schedule and the bond basis configurable; D-62 makes the reporting currency a setting; D-29 keeps credentials in Secrets, so the connection string comes from the environment fed by CloudNativePG's Secret, never a flag. Modeled by Claude Code for the core task: the author asked to see OpenFief's core running on the demo; the author has not reviewed this requirement yet.

## Source

> Configurable: worker concurrency, the end-of-day schedule and the bond basis (D-25); changing the basis revalues everything.

docs/DESIGN.md D-36
