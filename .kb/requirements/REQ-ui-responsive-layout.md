---
title: 'Every page of the built-in UI must work at phone width, 360 px and up, without scrolling sideways: the navigation folds behind a menu button, table rows become cards whose amounts stay monospaced, right-aligned and in their currency, charts resize, and forms span the screen with touch-sized targets.'
status: open
text_ref: docs/DESIGN.md#D-3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-3
  - lane:strict
semantic_text: 'Every page of the built-in UI must work at phone width, 360 px and up, without scrolling sideways: the navigation folds behind a menu button, table rows become cards whose amounts stay monospaced, right-aligned and in their currency, charts resize, and forms span the screen with touch-sized targets.'
logic_claims:
  - CLAIM-242A04FBFA9062A5
semantic_clauses:
  - 'Every page of the built-in UI must work at phone width, 360 px and up, without scrolling sideways: the navigation folds behind a menu button, table rows become cards whose amounts stay monospaced, right-aligned and in their currency, charts resize, and forms span the screen with touch-sized targets.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 797d48ce48fe835fe83146d8887cad025c6029d90139f8a7460fbca2364284e6
semantic_inventory:
  - claim_key: CLAIM-242A04FBFA9062A5
    claim_text: 'Every page of the built-in UI must work at phone width, 360 px and up, without scrolling sideways: the navigation folds behind a menu button, table rows become cards whose amounts stay monospaced, right-aligned and in their currency, charts resize, and forms span the screen with touch-sized targets'
    role: normative
    status: modeled
    span:
      start: 0
      end: 299
origin:
  kind: agent
  recorded_at: '2026-10-30T08:08:07.710Z'
id: REQ-ui-responsive-layout
type: req
---
Every page of the built-in UI must work at phone width, 360 px and up, without scrolling sideways: the navigation folds behind a menu button, table rows become cards whose amounts stay monospaced, right-aligned and in their currency, charts resize, and forms span the screen with touch-sized targets.

## Context

The author asked for a responsive design and a mobile PWA because the UI was not good on a phone. D-3 makes the built-in UI the default view; the design system sets 44 px touch targets and monospaced, right-aligned figures that always show their currency (D-20). Turning rows into cards rather than scrolling them is the orchestrating agent's choice for phones. Modeled by Claude Code for the author's request for a responsive UI and a mobile PWA; the author has not reviewed it yet.

## Source

> M0 ships two view plugins: a very simple built-in UI as the default, and Ghostfolio auto-push.

docs/DESIGN.md D-3

> Touch targets are at least 44 px.

docs/dev/design-system.md#space-and-shape
