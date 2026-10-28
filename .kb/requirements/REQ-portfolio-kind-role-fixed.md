---
title: The owner must be able to add sleeve kinds and rename any kind, and a stored kind's default role must never change.
status: open
text_ref: docs/DESIGN.md#GR-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-6
  - lane:strict
semantic_text: The owner must be able to add sleeve kinds and rename any kind, and a stored kind's default role must never change.
logic_claims:
  - CLAIM-F3C80F90416C618F
semantic_clauses:
  - The owner must be able to add sleeve kinds and rename any kind, and a stored kind's default role must never change.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8f24720422256d9f6b9a2846b9435c7b68cebe970303d6ea09aca781a3e42f4c
semantic_inventory:
  - claim_key: CLAIM-F3C80F90416C618F
    claim_text: The owner must be able to add sleeve kinds and rename any kind, and a stored kind's default role must never change
    role: normative
    status: modeled
    span:
      start: 0
      end: 114
origin:
  kind: agent
  recorded_at: '2026-10-28T09:35:05.953Z'
id: REQ-portfolio-kind-role-fixed
type: req
---
The owner must be able to add sleeve kinds and rename any kind, and a stored kind's default role must never change.

## Context

Author decision, sleeve kinds (given in chat to the orchestrator, not yet recorded in docs/DESIGN.md): role stays a fixed enum, defensive or offensive, because code enforces that agents may lease only from offensive sleeves (GR-6); a sleeve kind is a named category with a default role, from a built-in set the owner can extend; a sleeve has a kind, and its role defaults from the kind but is stored on the sleeve. The author allowed adding and renaming kinds. Fixing the default role is Claude Code's choice: sleeves store their own role, so changing a kind's default would either silently change nothing or leave sleeves that no longer match their kind, and turning a defensive kind offensive must never move capital into agents' reach (GR-6). The owner who wants another default creates a new kind. Modeled by Claude Code for the sleeve-kinds task; the author has not reviewed it yet.

## Source

> Role stays a fixed enum: defensive | offensive. Code enforces it: agents may only lease from offensive sleeves (GR-6), so roles are never owner-defined. Sleeve kind is new: a named category with a default role, from a built-in default set that the owner can extend. Defaults: emergency-fund (defensive), bonds (defensive), retirement (defensive; IKE/IKZE), core (defensive; long-term), satellite (offensive), agent-pool (offensive). A sleeve has a kind. Its role defaults from the kind but is stored on the sleeve. A default set of sleeves is created at install. The allocation chart (D-30) can group by kind.

author decision, sleeve kinds (chat); docs/DESIGN.md#GR-6, docs/DESIGN.md#D-30
