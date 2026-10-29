---
title: 'An install must start with three sleeves: bonds, core and satellite.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-71-5
  - lane:strict
text_ref: design-decisions:D-71#5
origin:
  kind: import
  ref: design-decisions:D-71#5
  recorded_at: '2026-10-29T05:36:36.910Z'
semantic_text: 'An install must start with three sleeves: bonds, core and satellite.'
logic_claims:
  - CLAIM-9209D3C8AAD97039
semantic_clauses:
  - 'An install must start with three sleeves: bonds, core and satellite.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: a8cbf0e7061aa370d101fc0570e5b25b7bf5564668298da6b23a7c27965a91e7
semantic_inventory:
  - claim_key: CLAIM-9209D3C8AAD97039
    claim_text: 'An install must start with three sleeves: bonds, core and satellite'
    role: normative
    status: modeled
    span:
      start: 0
      end: 67
id: REQ-portfolio-install-sleeves
type: req
---
An install must start with three sleeves: bonds, core and satellite.

## Context

Double-entry keeps cash, withholding tax and currency conversion consistent by construction, and it gives a later lease a natural home as a sub-account. The author recorded this decision with D-65 to D-70 and D-72 after the design session that answered the implementation's questions. REQ-portfolio-default-kinds and REQ-ledger-app-ensure-defaults refine it for the implementation: each missing default is created once, and a stored one is never changed.

## Source

> Every sleeve has a sleeve kind, which gives it a default role: defensive (never available to agents) or offensive (fiefs are leased from it, GR-6). The built-in kinds are agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive. The owner can add kinds and rename any kind, but a kind's default role never changes, and a sleeve whose role differs from its kind's needs the owner's explicit choice. An install starts with three sleeves: bonds, core and satellite.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-71#5
