---
title: The ledger's PostgreSQL schema must store sleeve kinds starting with the built-in ones, give every sleeve a stored kind, give sleeves stored before kinds existed the built-in kind whose default role is theirs, and refuse an offensive override on a sleeve that is not offensive and any change of a stored kind's default role.
status: open
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
semantic_text: The ledger's PostgreSQL schema must store sleeve kinds starting with the built-in ones, give every sleeve a stored kind, give sleeves stored before kinds existed the built-in kind whose default role is theirs, and refuse an offensive override on a sleeve that is not offensive and any change of a stored kind's default role.
logic_claims:
  - CLAIM-C68E34B74E7354FF
semantic_clauses:
  - The ledger's PostgreSQL schema must store sleeve kinds starting with the built-in ones, give every sleeve a stored kind, give sleeves stored before kinds existed the built-in kind whose default role is theirs, and refuse an offensive override on a sleeve that is not offensive and any change of a stored kind's default role.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 069158b39fb97555f7d4e8b90c7061ea3cd9b0037b3ee327227397fd96f7c890
semantic_inventory:
  - claim_key: CLAIM-C68E34B74E7354FF
    claim_text: The ledger's PostgreSQL schema must store sleeve kinds starting with the built-in ones, give every sleeve a stored kind, give sleeves stored before kinds existed the built-in kind whose default role is theirs, and refuse an offensive override on a sleeve that is not offensive and any change of a stored kind's default role
    role: normative
    status: modeled
    span:
      start: 0
      end: 323
origin:
  kind: agent
  recorded_at: '2026-10-28T09:53:38.341Z'
id: REQ-ledger-pg-sleeve-kinds
type: req
---
The ledger's PostgreSQL schema must store sleeve kinds starting with the built-in ones, give every sleeve a stored kind, give sleeves stored before kinds existed the built-in kind whose default role is theirs, and refuse an offensive override on a sleeve that is not offensive and any change of a stored kind's default role.

## Context

Author decision, sleeve kinds (given in chat to the orchestrator, not yet recorded in docs/DESIGN.md): role stays a fixed enum, defensive or offensive, because code enforces that agents may lease only from offensive sleeves (GR-6); a sleeve kind is a named category with a default role, from a built-in set the owner can extend; a sleeve has a kind, and its role defaults from the kind but is stored on the sleeve. Kinds are added to the ledger schema by a new migration, 0002, next to ledger.sleeves; schemas do not move. The database repeats the rules that keep capital out of agents' reach (GR-6) so that no other writer can break them. Refines REQ-ledger-pg-store. Modeled by Claude Code for the sleeve-kinds task; the author has not reviewed it yet.

## Source

> Role stays a fixed enum: defensive | offensive. Code enforces it: agents may only lease from offensive sleeves (GR-6), so roles are never owner-defined. Sleeve kind is new: a named category with a default role, from a built-in default set that the owner can extend. Defaults: emergency-fund (defensive), bonds (defensive), retirement (defensive; IKE/IKZE), core (defensive; long-term), satellite (offensive), agent-pool (offensive). A sleeve has a kind. Its role defaults from the kind but is stored on the sleeve. A default set of sleeves is created at install. The allocation chart (D-30) can group by kind.

author decision, sleeve kinds (chat); docs/DESIGN.md#GR-6, docs/DESIGN.md#D-30
