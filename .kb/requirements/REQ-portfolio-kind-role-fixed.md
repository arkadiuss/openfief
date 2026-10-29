---
title: The owner must be able to add sleeve kinds and rename any kind, and a stored kind's default role must never change.
status: open
text_ref: design-decisions:D-71#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-71-3
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
  kind: import
  ref: design-decisions:D-71#3
  recorded_at: '2026-10-29T05:35:06.980Z'
id: REQ-portfolio-kind-role-fixed
type: req
firmness: firm
required_evidence: asserted
---
The owner must be able to add sleeve kinds and rename any kind, and a stored kind's default role must never change.

## Context

Author decision, sleeve kinds (given in chat to the orchestrator, since recorded in docs/DESIGN.md as D-71): role stays a fixed enum, defensive or offensive, because code enforces that agents may lease only from offensive sleeves (GR-6); a sleeve kind is a named category with a default role, from a built-in set the owner can extend; a sleeve has a kind, and its role defaults from the kind but is stored on the sleeve. The author allowed adding and renaming kinds. Fixing the default role was first Claude Code's choice, which D-71 has since made the author's: sleeves store their own role, so changing a kind's default would either silently change nothing or leave sleeves that no longer match their kind, and turning a defensive kind offensive must never move capital into agents' reach (GR-6). The owner who wants another default creates a new kind. Modeled by Claude Code for the sleeve-kinds task; D-71 states it, so it is firm.

## Source

> Every sleeve has a sleeve kind, which gives it a default role: defensive (never available to agents) or offensive (fiefs are leased from it, GR-6). The built-in kinds are agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive. The owner can add kinds and rename any kind, but a kind's default role never changes, and a sleeve whose role differs from its kind's needs the owner's explicit choice. An install starts with three sleeves: bonds, core and satellite.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-71#3

> Role stays a fixed enum: defensive | offensive. Code enforces it: agents may only lease from offensive sleeves (GR-6), so roles are never owner-defined. Sleeve kind is new: a named category with a default role, from a built-in default set that the owner can extend. Defaults: emergency-fund (defensive), bonds (defensive), retirement (defensive; IKE/IKZE), core (defensive; long-term), satellite (offensive), agent-pool (offensive). A sleeve has a kind. Its role defaults from the kind but is stored on the sleeve. A default set of sleeves is created at install. The allocation chart (D-30) can group by kind.

author decision, sleeve kinds (chat, before D-71 was recorded); docs/DESIGN.md#GR-6, docs/DESIGN.md#D-30
