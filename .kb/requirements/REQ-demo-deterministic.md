---
title: 'The demo dataset must be a pure function of its seed and its today: the same seed and date must give the same instruments, statements, transactions, assertions, prices and index readings, with prices from integer arithmetic so that every value is exact.'
status: open
text_ref: docs/dev/testing.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-testing-md
  - lane:strict
semantic_text: 'The demo dataset must be a pure function of its seed and its today: the same seed and date must give the same instruments, statements, transactions, assertions, prices and index readings, with prices from integer arithmetic so that every value is exact.'
logic_claims:
  - CLAIM-A45D878A04C7A9CE
semantic_clauses:
  - 'The demo dataset must be a pure function of its seed and its today: the same seed and date must give the same instruments, statements, transactions, assertions, prices and index readings, with prices from integer arithmetic so that every value is exact.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0df422f02bc37de0aed455aee2c4ab4cf418857f113491fdcc9837c667b72ee9
semantic_inventory:
  - claim_key: CLAIM-A45D878A04C7A9CE
    claim_text: 'The demo dataset must be a pure function of its seed and its today: the same seed and date must give the same instruments, statements, transactions, assertions, prices and index readings, with prices from integer arithmetic so that every value is exact'
    role: normative
    status: modeled
    span:
      start: 0
      end: 252
origin:
  kind: agent
  recorded_at: '2026-10-28T09:31:16.269Z'
id: REQ-demo-deterministic
type: req
---
The demo dataset must be a pure function of its seed and its today: the same seed and date must give the same instruments, statements, transactions, assertions, prices and index readings, with prices from integer arithmetic so that every value is exact.

## Context

A demo that changes on every run can't be reproduced or debugged, and docs/dev/testing.md forbids unseeded randomness and the wall clock; the date is injected and the generator is seeded. No float64 touches an amount (AGENTS.md). Modeled by Claude Code for the demo-data task: the author asked to see OpenFief on the Hetzner demo with real pipelines on fake data; the author has not reviewed this requirement yet.

## Source

> No unseeded randomness, no network in unit or replay tests, no dependence on test order.

docs/dev/testing.md
