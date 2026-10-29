---
title: The built-in sleeve kinds must be agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-71-2
  - lane:strict
text_ref: design-decisions:D-71#2
origin:
  kind: import
  ref: design-decisions:D-71#2
  recorded_at: '2026-10-29T05:34:45.098Z'
semantic_text: The built-in sleeve kinds must be agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive.
logic_claims:
  - CLAIM-4AF56B2D5D413AC8
semantic_clauses:
  - The built-in sleeve kinds must be agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 4f0f42b24f17ac110f4bff4daf665e601bcafced3799c7c571c38d99a9f9daca
semantic_inventory:
  - claim_key: CLAIM-4AF56B2D5D413AC8
    claim_text: The built-in sleeve kinds must be agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive
    role: normative
    status: modeled
    span:
      start: 0
      end: 148
id: REQ-portfolio-builtin-kinds
type: req
---
The built-in sleeve kinds must be agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive.

## Context

Double-entry keeps cash, withholding tax and currency conversion consistent by construction, and it gives a later lease a natural home as a sub-account. The author recorded this decision with D-65 to D-70 and D-72 after the design session that answered the implementation's questions. REQ-portfolio-default-kinds refines it for the implementation, with the kinds' keys and how the defaults are created; it is not promoted because it adds how the defaults are stored, which D-71 does not say.

## Source

> Every sleeve has a sleeve kind, which gives it a default role: defensive (never available to agents) or offensive (fiefs are leased from it, GR-6). The built-in kinds are agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive. The owner can add kinds and rename any kind, but a kind's default role never changes, and a sleeve whose role differs from its kind's needs the owner's explicit choice. An install starts with three sleeves: bonds, core and satellite.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-71#2
