---
title: A sleeve whose role differs from its kind's default role must need the owner's explicit choice.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-71-4
  - lane:strict
text_ref: design-decisions:D-71#4
origin:
  kind: import
  ref: design-decisions:D-71#4
  recorded_at: '2026-10-29T05:35:53.118Z'
semantic_text: A sleeve whose role differs from its kind's default role must need the owner's explicit choice.
logic_claims:
  - CLAIM-310D771A247974AC
semantic_clauses:
  - A sleeve whose role differs from its kind's default role must need the owner's explicit choice.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 976d1542c81d33343d25575486651c8924cc0fa3ce0e6633d23e6f22f1ef230b
semantic_inventory:
  - claim_key: CLAIM-310D771A247974AC
    claim_text: A sleeve whose role differs from its kind's default role must need the owner's explicit choice
    role: normative
    status: modeled
    span:
      start: 0
      end: 94
id: REQ-portfolio-role-differs-explicit
type: req
---
A sleeve whose role differs from its kind's default role must need the owner's explicit choice.

## Context

Double-entry keeps cash, withholding tax and currency conversion consistent by construction, and it gives a later lease a natural home as a sub-account. The author recorded this decision with D-65 to D-70 and D-72 after the design session that answered the implementation's questions. REQ-portfolio-sleeve-role-override is not promoted for it because it differs: it asks for a stored explicit override only when a sleeve is offensive in a defensive kind and lets a sleeve be defensive in an offensive kind without one, while D-71 asks for the owner's explicit choice whenever the roles differ. Whether the owner choosing the role is that explicit choice is for the author to confirm.

## Source

> Every sleeve has a sleeve kind, which gives it a default role: defensive (never available to agents) or offensive (fiefs are leased from it, GR-6). The built-in kinds are agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive. The owner can add kinds and rename any kind, but a kind's default role never changes, and a sleeve whose role differs from its kind's needs the owner's explicit choice. An install starts with three sleeves: bonds, core and satellite.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-71#4
