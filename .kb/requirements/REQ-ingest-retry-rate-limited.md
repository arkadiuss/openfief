---
title: The core must refuse a retry of a connector instance's sync less than a minute after the last one it started.
status: open
text_ref: design-decisions:D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196
  - lane:strict
semantic_text: The core must refuse a retry of a connector instance's sync less than a minute after the last one it started.
logic_claims:
  - CLAIM-CDE2FFDE3B9D78D0
semantic_clauses:
  - The core must refuse a retry of a connector instance's sync less than a minute after the last one it started.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c2fc957a8d8157dfbb8103377c7173b896051089fbc5f98754fb650bb9bda701
semantic_inventory:
  - claim_key: CLAIM-CDE2FFDE3B9D78D0
    claim_text: The core must refuse a retry of a connector instance's sync less than a minute after the last one it started
    role: normative
    status: modeled
    span:
      start: 0
      end: 108
origin:
  kind: agent
  recorded_at: '2026-11-06T04:23:06.377Z'
id: REQ-ingest-retry-rate-limited
type: req
---
The core must refuse a retry of a connector instance's sync less than a minute after the last one it started.

## Context

The design session asked that retries be rate-limited per instance, so a button pressed repeatedly can't hammer a broker's API; a minute is this change's choice. Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-196]** ... An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place ...

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196
