---
title: Ensuring the defaults must store every missing built-in sleeve kind and default sleeve in one database transaction, never change a stored kind or sleeve, and store nothing more when run again or by several processes at once.
status: open
text_ref: docs/DESIGN.md#GR-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-6
  - lane:strict
semantic_text: Ensuring the defaults must store every missing built-in sleeve kind and default sleeve in one database transaction, never change a stored kind or sleeve, and store nothing more when run again or by several processes at once.
logic_claims:
  - CLAIM-F5D735EC1994D2DF
semantic_clauses:
  - Ensuring the defaults must store every missing built-in sleeve kind and default sleeve in one database transaction, never change a stored kind or sleeve, and store nothing more when run again or by several processes at once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f65a57081f6bc07670201dc726052cb071070ab1f89ef97ca53010b21c5ba16e
semantic_inventory:
  - claim_key: CLAIM-F5D735EC1994D2DF
    claim_text: Ensuring the defaults must store every missing built-in sleeve kind and default sleeve in one database transaction, never change a stored kind or sleeve, and store nothing more when run again or by several processes at once
    role: normative
    status: modeled
    span:
      start: 0
      end: 223
origin:
  kind: agent
  recorded_at: '2026-10-28T09:53:27.076Z'
id: REQ-ledger-app-ensure-defaults
type: req
---
Ensuring the defaults must store every missing built-in sleeve kind and default sleeve in one database transaction, never change a stored kind or sleeve, and store nothing more when run again or by several processes at once.

## Context

Author decision, sleeve kinds (given in chat to the orchestrator, not yet recorded in docs/DESIGN.md): role stays a fixed enum, defensive or offensive, because code enforces that agents may lease only from offensive sleeves (GR-6); a sleeve kind is a named category with a default role, from a built-in set the owner can extend; a sleeve has a kind, and its role defaults from the kind but is stored on the sleeve. The author asked for a small default set of sleeves created at install, idempotently. Running it on every start is safe, so the core does not need to remember whether it was installed. Refines REQ-portfolio-default-kinds. Modeled by Claude Code for the sleeve-kinds task; the author has not reviewed it yet.

## Source

> Role stays a fixed enum: defensive | offensive. Code enforces it: agents may only lease from offensive sleeves (GR-6), so roles are never owner-defined. Sleeve kind is new: a named category with a default role, from a built-in default set that the owner can extend. Defaults: emergency-fund (defensive), bonds (defensive), retirement (defensive; IKE/IKZE), core (defensive; long-term), satellite (offensive), agent-pool (offensive). A sleeve has a kind. Its role defaults from the kind but is stored on the sleeve. A default set of sleeves is created at install. The allocation chart (D-30) can group by kind.

author decision, sleeve kinds (chat); docs/DESIGN.md#GR-6, docs/DESIGN.md#D-30
