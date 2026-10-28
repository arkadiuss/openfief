---
title: The ledger must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role.
status: open
text_ref: docs/DESIGN.md#GR-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-6
  - lane:strict
semantic_text: The ledger must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role.
logic_claims:
  - CLAIM-600720DA21BCDE78
semantic_clauses:
  - The ledger must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 199c684de48a302355d2df39460b341dc2c820ca7db901225b191748a5fa9331
semantic_inventory:
  - claim_key: CLAIM-600720DA21BCDE78
    claim_text: The ledger must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
origin:
  kind: agent
  recorded_at: '2026-10-28T09:53:17.239Z'
id: REQ-ledger-app-sleeve-kinds
type: req
---
The ledger must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role.

## Context

Author decision, sleeve kinds (given in chat to the orchestrator, not yet recorded in docs/DESIGN.md): role stays a fixed enum, defensive or offensive, because code enforces that agents may lease only from offensive sleeves (GR-6); a sleeve kind is a named category with a default role, from a built-in set the owner can extend; a sleeve has a kind, and its role defaults from the kind but is stored on the sleeve. Sleeves and their kinds are stored by the ledger context because the portfolio context has no store in M0; the rules themselves are portfolio.CheckKind and portfolio.CheckSleeve. Refines REQ-portfolio-sleeve-kind, REQ-portfolio-kind-role-fixed and REQ-portfolio-sleeve-role-override for the use cases the UI and the read API call. Modeled by Claude Code for the sleeve-kinds task; the author has not reviewed it yet.

## Source

> Role stays a fixed enum: defensive | offensive. Code enforces it: agents may only lease from offensive sleeves (GR-6), so roles are never owner-defined. Sleeve kind is new: a named category with a default role, from a built-in default set that the owner can extend. Defaults: emergency-fund (defensive), bonds (defensive), retirement (defensive; IKE/IKZE), core (defensive; long-term), satellite (offensive), agent-pool (offensive). A sleeve has a kind. Its role defaults from the kind but is stored on the sleeve. A default set of sleeves is created at install. The allocation chart (D-30) can group by kind.

author decision, sleeve kinds (chat); docs/DESIGN.md#GR-6, docs/DESIGN.md#D-30
