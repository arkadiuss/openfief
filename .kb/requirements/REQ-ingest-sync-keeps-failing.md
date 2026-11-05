---
title: A connector instance's sync must count as one that keeps failing when at least three attempts in a row have failed since its latest success and they span at least fifteen minutes from the first one's start to the latest one's end, and one success must clear it.
status: open
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
semantic_text: A connector instance's sync must count as one that keeps failing when at least three attempts in a row have failed since its latest success and they span at least fifteen minutes from the first one's start to the latest one's end, and one success must clear it.
logic_claims:
  - CLAIM-32DF80F61CB1CFF4
semantic_clauses:
  - A connector instance's sync must count as one that keeps failing when at least three attempts in a row have failed since its latest success and they span at least fifteen minutes from the first one's start to the latest one's end, and one success must clear it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 37bf9f3243ca0af1996c2ab8a38474b38eb6ae0a581ede538276a9c4509b53c6
semantic_inventory:
  - claim_key: CLAIM-32DF80F61CB1CFF4
    claim_text: A connector instance's sync must count as one that keeps failing when at least three attempts in a row have failed since its latest success and they span at least fifteen minutes from the first one's start to the latest one's end, and one success must clear it
    role: normative
    status: modeled
    span:
      start: 0
      end: 260
origin:
  kind: agent
  recorded_at: '2026-11-05T23:39:12.216Z'
id: REQ-ingest-sync-keeps-failing
type: req
---
A connector instance's sync must count as one that keeps failing when at least three attempts in a row have failed since its latest success and they span at least fifteen minutes from the first one's start to the latest one's end, and one success must clear it.

## Context

D-196 names a sync that keeps failing but not when it does; this is the core's reading, so a plugin's restart or a short outage is retried without bothering the owner. With the core's default retries (5 minutes, doubling) the third failure ends about fifteen minutes after the first began. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> **[D-196]** ... A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196
