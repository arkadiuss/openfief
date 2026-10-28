---
title: The demo loader must refuse, before writing anything, a database that holds ledger or valuation data and has no demo marker, with or without --reset-demo; it may load an empty database, which it first marks as a demo database, and may reset only a database that carries the marker.
status: open
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
semantic_text: The demo loader must refuse, before writing anything, a database that holds ledger or valuation data and has no demo marker, with or without --reset-demo; it may load an empty database, which it first marks as a demo database, and may reset only a database that carries the marker.
logic_claims:
  - CLAIM-276A49FAB9D7BECB
semantic_clauses:
  - The demo loader must refuse, before writing anything, a database that holds ledger or valuation data and has no demo marker, with or without --reset-demo; it may load an empty database, which it first marks as a demo database, and may reset only a database that carries the marker.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 05188f29951cb8723e4207d0e022cbb38377928198de5d5799b15471d3a87cf4
semantic_inventory:
  - claim_key: CLAIM-276A49FAB9D7BECB
    claim_text: The demo loader must refuse, before writing anything, a database that holds ledger or valuation data and has no demo marker, with or without --reset-demo; it may load an empty database, which it first marks as a demo database, and may reset only a database that carries the marker
    role: normative
    status: modeled
    span:
      start: 0
      end: 280
origin:
  kind: agent
  recorded_at: '2026-10-28T09:31:04.801Z'
id: REQ-demo-guard-refuses-real-data
type: req
---
The demo loader must refuse, before writing anything, a database that holds ledger or valuation data and has no demo marker, with or without --reset-demo; it may load an empty database, which it first marks as a demo database, and may reset only a database that carries the marker.

## Context

The demo loader writes and drops whole schemas, so pointed at a real OpenFief database by mistake it could destroy a portfolio; the author asked that this be impossible by accident. The marker is a row only the loader writes, in its own demo schema, and the check runs under an advisory lock before any migration. Exercises GR-12 and AGENTS.md's no-real-data rule. Modeled by Claude Code for the demo-data task: the author asked to see OpenFief on the Hetzner demo with real pipelines on fake data; the author has not reviewed this requirement yet.

## Source

> The privacy and security model comes first, designed per use case.

docs/DESIGN.md GR-12; AGENTS.md: "No real financial data, account numbers or credentials anywhere: code, tests, recordings, logs, commits."
