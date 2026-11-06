---
title: Every connector sync attempt must be recorded in an append-only history with its batch, mode, start, end and next sync, and a failed one with how many attempts in a row have failed and an error class taken from the known errors in its chain, never the error's text.
status: open
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
semantic_text: Every connector sync attempt must be recorded in an append-only history with its batch, mode, start, end and next sync, and a failed one with how many attempts in a row have failed and an error class taken from the known errors in its chain, never the error's text.
logic_claims:
  - CLAIM-0B1092DF451DC233
semantic_clauses:
  - Every connector sync attempt must be recorded in an append-only history with its batch, mode, start, end and next sync, and a failed one with how many attempts in a row have failed and an error class taken from the known errors in its chain, never the error's text.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1a8af8688614678d4a398e4b954b2ea7bab7757538d81860e98527688118d638
semantic_inventory:
  - claim_key: CLAIM-0B1092DF451DC233
    claim_text: Every connector sync attempt must be recorded in an append-only history with its batch, mode, start, end and next sync, and a failed one with how many attempts in a row have failed and an error class taken from the known errors in its chain, never the error's text
    role: normative
    status: modeled
    span:
      start: 0
      end: 264
origin:
  kind: agent
  recorded_at: '2026-11-05T23:39:06.970Z'
id: REQ-ingest-sync-attempt-history
type: req
---
Every connector sync attempt must be recorded in an append-only history with its batch, mode, start, end and next sync, and a failed one with how many attempts in a row have failed and an error class taken from the known errors in its chain, never the error's text.

## Context

D-196 has a connector sync that keeps failing raise an item that shows what happened, and D-193 puts any failing sync in the overview's status sentence. The core owns syncs and their retries (D-49). GR-12 forbids keeping a value from an error, so only a class is kept. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> **[D-196]** ... A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196

> **[D-49]** The core calls connectors: it owns schedules, retries, import batches and watermarks.

docs/DESIGN.md#D-49

> **[GR-12]** Never log amounts, holdings, account numbers or tokens.

docs/DESIGN.md#GR-12
