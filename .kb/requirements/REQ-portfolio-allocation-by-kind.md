---
title: Each sleeve's kind must be available by sleeve ID so that the allocation chart can group the portfolio by sleeve kind.
status: open
text_ref: docs/DESIGN.md#D-30
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-30
  - lane:strict
semantic_text: Each sleeve's kind must be available by sleeve ID so that the allocation chart can group the portfolio by sleeve kind.
logic_claims:
  - CLAIM-08EB2CE9A22A4F38
semantic_clauses:
  - Each sleeve's kind must be available by sleeve ID so that the allocation chart can group the portfolio by sleeve kind.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c77275e7475568d02fd334723566deb3885b617d6f0cba179fc83d5e1ee20986
semantic_inventory:
  - claim_key: CLAIM-08EB2CE9A22A4F38
    claim_text: Each sleeve's kind must be available by sleeve ID so that the allocation chart can group the portfolio by sleeve kind
    role: normative
    status: modeled
    span:
      start: 0
      end: 117
origin:
  kind: agent
  recorded_at: '2026-10-28T09:35:44.586Z'
id: REQ-portfolio-allocation-by-kind
type: req
---
Each sleeve's kind must be available by sleeve ID so that the allocation chart can group the portfolio by sleeve kind.

## Context

Author decision, sleeve kinds (given in chat to the orchestrator, not yet recorded in docs/DESIGN.md): role stays a fixed enum, defensive or offensive, because code enforces that agents may lease only from offensive sleeves (GR-6); a sleeve kind is a named category with a default role, from a built-in set the owner can extend; a sleeve has a kind, and its role defaults from the kind but is stored on the sleeve. D-30 groups the allocation chart by asset class, currency, broker or sleeve; the author added grouping by sleeve kind. The read API resolves an account's sleeve to its kind with the portfolio function. Modeled by Claude Code for the sleeve-kinds task; the author has not reviewed it yet.

## Source

> Role stays a fixed enum: defensive | offensive. Code enforces it: agents may only lease from offensive sleeves (GR-6), so roles are never owner-defined. Sleeve kind is new: a named category with a default role, from a built-in default set that the owner can extend. Defaults: emergency-fund (defensive), bonds (defensive), retirement (defensive; IKE/IKZE), core (defensive; long-term), satellite (offensive), agent-pool (offensive). A sleeve has a kind. Its role defaults from the kind but is stored on the sleeve. A default set of sleeves is created at install. The allocation chart (D-30) can group by kind.

author decision, sleeve kinds (chat); docs/DESIGN.md#GR-6, docs/DESIGN.md#D-30
