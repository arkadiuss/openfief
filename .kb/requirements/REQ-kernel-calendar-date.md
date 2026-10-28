---
title: A date must be a calendar date with no time of day or time zone, read and written as ISO 8601 YYYY-MM-DD, and a date that does not exist must be rejected.
status: open
text_ref: docs/DESIGN.md#D-21
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-21
  - lane:strict
semantic_text: A date must be a calendar date with no time of day or time zone, read and written as ISO 8601 YYYY-MM-DD, and a date that does not exist must be rejected.
logic_claims:
  - CLAIM-28F0AA3D40A8163E
semantic_clauses:
  - A date must be a calendar date with no time of day or time zone, read and written as ISO 8601 YYYY-MM-DD, and a date that does not exist must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1ec8cdabcbbdeb10f7a3ea51283da1201c3484046065483c197612f35928a2d2
semantic_inventory:
  - claim_key: CLAIM-28F0AA3D40A8163E
    claim_text: A date must be a calendar date with no time of day or time zone, read and written as ISO 8601 YYYY-MM-DD, and a date that does not exist must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 153
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:59.053Z'
id: REQ-kernel-calendar-date
type: req
---
A date must be a calendar date with no time of day or time zone, read and written as ISO 8601 YYYY-MM-DD, and a date that does not exist must be rejected.

## Context

Trade dates, lot purchase dates (D-9, D-23) and end-of-day prices (D-21) are calendar dates; carrying a time zone would move them by a day around midnight. Domain code receives dates as parameters and never reads the wall clock. Rejecting 2026-02-30 instead of normalizing it to March stops a bad import row from landing on a wrong date. Modeled by Claude Code for the M0 foundation task (shared kernel); the author has not reviewed it yet.

## Source

> Prices are end of day in M0.

docs/DESIGN.md D-21

> No wall clock: inject the date.

docs/dev/testing.md#stability
