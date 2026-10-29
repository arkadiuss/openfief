---
title: The built-in UI must take every color, font, mark and mascot from the design system in design/, served by the UI itself, write no color outside design/tokens.css, and load the design system's fonts and tokens on every page.
status: open
text_ref: docs/dev/design-system.md#how-it-works
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-design-system-md-how-it-works
  - lane:strict
semantic_text: The built-in UI must take every color, font, mark and mascot from the design system in design/, served by the UI itself, write no color outside design/tokens.css, and load the design system's fonts and tokens on every page.
logic_claims:
  - CLAIM-F3B942E6EEF9AD78
semantic_clauses:
  - The built-in UI must take every color, font, mark and mascot from the design system in design/, served by the UI itself, write no color outside design/tokens.css, and load the design system's fonts and tokens on every page.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 873285bfe9f5b77c1c8aafce5b261c05258a315d6b63c41b0370867d0164ed8a
semantic_inventory:
  - claim_key: CLAIM-F3B942E6EEF9AD78
    claim_text: The built-in UI must take every color, font, mark and mascot from the design system in design/, served by the UI itself, write no color outside design/tokens.css, and load the design system's fonts and tokens on every page
    role: normative
    status: modeled
    span:
      start: 0
      end: 222
origin:
  kind: agent
  recorded_at: '2026-10-29T09:49:53.149Z'
id: REQ-ui-design-system
type: req
---
The built-in UI must take every color, font, mark and mascot from the design system in design/, served by the UI itself, write no color outside design/tokens.css, and load the design system's fonts and tokens on every page.

## Context

The author's design system makes design/ the single source for the docs site and the built-in UI; its rule book left wiring the UI to it open, marked (P), between an embedding Go package and a checked copy. The UI agent picked the Go package design/design.go, so nothing is copied. Requested by the orchestrator for the UI port.

## Source

> design/ is the single source; the docs site and the built-in UI consume it and never copy a value out of it.

docs/dev/design-system.md#how-it-works
