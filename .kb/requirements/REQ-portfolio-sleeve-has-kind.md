---
title: 'Every sleeve must have a sleeve kind, which gives it a default role: defensive, never available to agents, or offensive, the part fiefs are leased from.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-71-1
  - lane:strict
text_ref: design-decisions:D-71#1
origin:
  kind: import
  ref: design-decisions:D-71#1
  recorded_at: '2026-10-29T05:33:54.990Z'
semantic_text: 'Every sleeve must have a sleeve kind, which gives it a default role: defensive, never available to agents, or offensive, the part fiefs are leased from.'
logic_claims:
  - CLAIM-453D2B8C75DB8333
semantic_clauses:
  - 'Every sleeve must have a sleeve kind, which gives it a default role: defensive, never available to agents, or offensive, the part fiefs are leased from.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: fb1ee0388b3152788658032da22c3bc2949bb001487cb9073734b374a7d3413c
semantic_inventory:
  - claim_key: CLAIM-453D2B8C75DB8333
    claim_text: 'Every sleeve must have a sleeve kind, which gives it a default role: defensive, never available to agents, or offensive, the part fiefs are leased from'
    role: normative
    status: modeled
    span:
      start: 0
      end: 151
id: REQ-portfolio-sleeve-has-kind
type: req
---
Every sleeve must have a sleeve kind, which gives it a default role: defensive, never available to agents, or offensive, the part fiefs are leased from.

## Context

Double-entry keeps cash, withholding tax and currency conversion consistent by construction, and it gives a later lease a natural home as a sub-account. GR-6 leases each agent's budget from the offensive part of the portfolio only. The author recorded this decision with D-65 to D-70 and D-72 after the design session that answered the implementation's questions; REQ-portfolio-sleeve-kind and REQ-portfolio-sleeve-role refine it for the implementation.

## Source

> Every sleeve has a sleeve kind, which gives it a default role: defensive (never available to agents) or offensive (fiefs are leased from it, GR-6). The built-in kinds are agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive. The owner can add kinds and rename any kind, but a kind's default role never changes, and a sleeve whose role differs from its kind's needs the owner's explicit choice. An install starts with three sleeves: bonds, core and satellite.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-71#1
