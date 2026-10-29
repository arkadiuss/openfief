---
title: 'The core must serve the owner''s sleeve entries as one service to view plugins acting for an owner session: adding a sleeve kind with a fixed default role, renaming a kind, adding a sleeve with an ID made from its name, and changing a sleeve''s name, kind or role, refusing a taken key or ID, an unknown kind or sleeve and a role other than the kind''s default without the owner''s override.'
status: open
text_ref: docs/DESIGN.md#D-79
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-79
  - lane:strict
semantic_text: 'The core must serve the owner''s sleeve entries as one service to view plugins acting for an owner session: adding a sleeve kind with a fixed default role, renaming a kind, adding a sleeve with an ID made from its name, and changing a sleeve''s name, kind or role, refusing a taken key or ID, an unknown kind or sleeve and a role other than the kind''s default without the owner''s override.'
logic_claims:
  - CLAIM-228501B16D7BBFDB
semantic_clauses:
  - 'The core must serve the owner''s sleeve entries as one service to view plugins acting for an owner session: adding a sleeve kind with a fixed default role, renaming a kind, adding a sleeve with an ID made from its name, and changing a sleeve''s name, kind or role, refusing a taken key or ID, an unknown kind or sleeve and a role other than the kind''s default without the owner''s override.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7ec67c06e193023b3247865035cd72a0f9f470a1f67bbaf8e7db0699345dc51e
semantic_inventory:
  - claim_key: CLAIM-228501B16D7BBFDB
    claim_text: 'The core must serve the owner''s sleeve entries as one service to view plugins acting for an owner session: adding a sleeve kind with a fixed default role, renaming a kind, adding a sleeve with an ID made from its name, and changing a sleeve''s name, kind or role, refusing a taken key or ID, an unknown kind or sleeve and a role other than the kind''s default without the owner''s override'
    role: normative
    status: modeled
    span:
      start: 0
      end: 386
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:55.270Z'
id: REQ-portfolio-sleeve-service
type: req
---
The core must serve the owner's sleeve entries as one service to view plugins acting for an owner session: adding a sleeve kind with a fixed default role, renaming a kind, adding a sleeve with an ID made from its name, and changing a sleeve's name, kind or role, refusing a taken key or ID, an unknown kind or sleeve and a role other than the kind's default without the owner's override.

## Context

D-71 lets the owner add and rename kinds and choose a sleeve's role explicitly; D-73 has the wizard create sleeves; D-79 makes each kind of owner entry its own owner-only service called only by a view plugin. Refines REQ-core-owner-entry-services and REQ-portfolio-kind-role-fixed. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-71]** Every sleeve has a sleeve kind, which gives it a default role: defensive (never available to agents) or offensive (fiefs are leased from it, GR-6). The built-in kinds are agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive. The owner can add kinds and rename any kind, but a kind's default role never changes, and a sleeve whose role differs from its kind's needs the owner's explicit choice. An install starts with three sleeves: bonds, core and satellite.

docs/DESIGN.md#D-71

> **[D-79]** What the owner enters by hand reaches the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind. Each is owner-only, versioned on its own, and called only by a view plugin acting for an owner session (D-69). The read API (D-6) and script tokens (D-27) stay read-only.

docs/DESIGN.md#D-79
