---
title: A sleeve's role must default to its kind's default role; a sleeve may be defensive in an offensive kind, and may be offensive in a defensive kind only with the owner's explicit offensive override, which no other sleeve may carry.
status: open
text_ref: docs/DESIGN.md#GR-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-6
  - lane:strict
semantic_text: A sleeve's role must default to its kind's default role; a sleeve may be defensive in an offensive kind, and may be offensive in a defensive kind only with the owner's explicit offensive override, which no other sleeve may carry.
logic_claims:
  - CLAIM-59010602175AB516
semantic_clauses:
  - A sleeve's role must default to its kind's default role; a sleeve may be defensive in an offensive kind, and may be offensive in a defensive kind only with the owner's explicit offensive override, which no other sleeve may carry.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f6e0b5a7c3595de99fd3e779e391c587c4af6dfded95d640936cd97a994fb651
semantic_inventory:
  - claim_key: CLAIM-59010602175AB516
    claim_text: A sleeve's role must default to its kind's default role; a sleeve may be defensive in an offensive kind, and may be offensive in a defensive kind only with the owner's explicit offensive override, which no other sleeve may carry
    role: normative
    status: modeled
    span:
      start: 0
      end: 228
origin:
  kind: agent
  recorded_at: '2026-10-28T09:35:20.432Z'
id: REQ-portfolio-sleeve-role-override
type: req
---
A sleeve's role must default to its kind's default role; a sleeve may be defensive in an offensive kind, and may be offensive in a defensive kind only with the owner's explicit offensive override, which no other sleeve may carry.

## Context

Author decision, sleeve kinds (given in chat to the orchestrator, not yet recorded in docs/DESIGN.md): role stays a fixed enum, defensive or offensive, because code enforces that agents may lease only from offensive sleeves (GR-6); a sleeve kind is a named category with a default role, from a built-in set the owner can extend; a sleeve has a kind, and its role defaults from the kind but is stored on the sleeve. The author asked to choose whether a sleeve may override its kind's role, noting that offensive to defensive is always safe and defensive to offensive should be allowed only explicitly. Offensive to defensive only takes capital out of agents' reach. Defensive to offensive makes capital leasable to agents (GR-6), so it needs a stored, explicit flag that the read API and audit can show. Refines REQ-portfolio-sleeve-role. Modeled by Claude Code for the sleeve-kinds task; the author has not reviewed it yet.

## Source

> Role stays a fixed enum: defensive | offensive. Code enforces it: agents may only lease from offensive sleeves (GR-6), so roles are never owner-defined. Sleeve kind is new: a named category with a default role, from a built-in default set that the owner can extend. Defaults: emergency-fund (defensive), bonds (defensive), retirement (defensive; IKE/IKZE), core (defensive; long-term), satellite (offensive), agent-pool (offensive). A sleeve has a kind. Its role defaults from the kind but is stored on the sleeve. A default set of sleeves is created at install. The allocation chart (D-30) can group by kind.

author decision, sleeve kinds (chat); docs/DESIGN.md#GR-6, docs/DESIGN.md#D-30

> Each thesis agent has a budget leased from the offensive part of the portfolio, with strong guarantees that the agent can never get more of the portfolio than its lease.

docs/DESIGN.md GR-6
