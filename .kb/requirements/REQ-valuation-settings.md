---
title: Worker concurrency, the daily end-of-day time with its time zone and the bond basis must be configurable, and invalid settings must be rejected.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: Worker concurrency, the daily end-of-day time with its time zone and the bond basis must be configurable, and invalid settings must be rejected.
logic_claims:
  - CLAIM-330FF024C5FA76C7
semantic_clauses:
  - Worker concurrency, the daily end-of-day time with its time zone and the bond basis must be configurable, and invalid settings must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 51fc61fe582c5fe4bc9ecdecc4acb4840ce73ea0118b2304dd548f9effda3434
semantic_inventory:
  - claim_key: CLAIM-330FF024C5FA76C7
    claim_text: Worker concurrency, the daily end-of-day time with its time zone and the bond basis must be configurable, and invalid settings must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 143
origin:
  kind: agent
  recorded_at: '2026-10-28T07:58:47.561Z'
id: REQ-valuation-settings
type: req
---
Worker concurrency, the daily end-of-day time with its time zone and the bond basis must be configurable, and invalid settings must be rejected.

## Context

D-36 lists these three as configurable. The end-of-day schedule is a daily time in a time zone, run as a River periodic job; River has no cron parser of its own and a daily time needs no extra dependency. Refines REQ-decision-d36-8-revaluation-settings. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Configurable: worker concurrency, the end-of-day schedule and the bond basis (D-25); changing the basis revalues everything.

docs/DESIGN.md D-36
