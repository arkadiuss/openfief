---
title: Every sleeve must have exactly one sleeve kind among the stored kinds, and a sleeve kind must have a key of lowercase hyphenated words, a name and a default role, defensive or offensive.
status: open
text_ref: docs/DESIGN.md#GR-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-6
  - lane:strict
semantic_text: Every sleeve must have exactly one sleeve kind among the stored kinds, and a sleeve kind must have a key of lowercase hyphenated words, a name and a default role, defensive or offensive.
logic_claims:
  - CLAIM-F1747C7DCD9862EE
semantic_clauses:
  - Every sleeve must have exactly one sleeve kind among the stored kinds, and a sleeve kind must have a key of lowercase hyphenated words, a name and a default role, defensive or offensive.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a947b83fcac5a7e249351ae21dd4cf4f30459e2cb974eac55e758117b1fce3d6
semantic_inventory:
  - claim_key: CLAIM-F1747C7DCD9862EE
    claim_text: Every sleeve must have exactly one sleeve kind among the stored kinds, and a sleeve kind must have a key of lowercase hyphenated words, a name and a default role, defensive or offensive
    role: normative
    status: modeled
    span:
      start: 0
      end: 185
origin:
  kind: agent
  recorded_at: '2026-10-28T09:34:52.161Z'
id: REQ-portfolio-sleeve-kind
type: req
---
Every sleeve must have exactly one sleeve kind among the stored kinds, and a sleeve kind must have a key of lowercase hyphenated words, a name and a default role, defensive or offensive.

## Context

Author decision, sleeve kinds (given in chat to the orchestrator, not yet recorded in docs/DESIGN.md): role stays a fixed enum, defensive or offensive, because code enforces that agents may lease only from offensive sleeves (GR-6); a sleeve kind is a named category with a default role, from a built-in set the owner can extend; a sleeve has a kind, and its role defaults from the kind but is stored on the sleeve. Kind keys are restricted to lowercase hyphenated words up to 64 characters so they are stable identifiers in URLs and the read API; the restriction is Claude Code's choice. Modeled by Claude Code for the sleeve-kinds task; the author has not reviewed it yet.

## Source

> Role stays a fixed enum: defensive | offensive. Code enforces it: agents may only lease from offensive sleeves (GR-6), so roles are never owner-defined. Sleeve kind is new: a named category with a default role, from a built-in default set that the owner can extend. Defaults: emergency-fund (defensive), bonds (defensive), retirement (defensive; IKE/IKZE), core (defensive; long-term), satellite (offensive), agent-pool (offensive). A sleeve has a kind. Its role defaults from the kind but is stored on the sleeve. A default set of sleeves is created at install. The allocation chart (D-30) can group by kind.

author decision, sleeve kinds (chat); docs/DESIGN.md#GR-6, docs/DESIGN.md#D-30
