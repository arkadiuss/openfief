---
title: The core must read its database connection string from OPENFIEF_DSN and its listen address, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout from flags with documented defaults, and refuse to start on a missing or invalid value.
status: closed
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

Closed: superseded by REQ-core-serve-flags. The bond basis flag is gone with D-77. Modeled by Claude Code for the values work package; the author has not reviewed it yet.
