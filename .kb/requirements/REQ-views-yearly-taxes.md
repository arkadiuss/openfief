---
title: The read API must list, for a year, the yearly tax of each tax wrapper taxed on its balance, computed from the summed daily gross totals with cash of the wrapper's accounts only under a law in force, and otherwise saying it is not in force.
status: closed
text_ref: docs/DESIGN.md#D-76
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-76
  - lane:strict
semantic_text: The read API must list, for a year, the yearly tax of each tax wrapper taxed on its balance, computed from the summed daily gross totals with cash of the wrapper's accounts only under a law in force, and otherwise saying it is not in force.
logic_claims:
  - CLAIM-3D430515651915F7
semantic_clauses:
  - The read API must list, for a year, the yearly tax of each tax wrapper taxed on its balance, computed from the summed daily gross totals with cash of the wrapper's accounts only under a law in force, and otherwise saying it is not in force.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c78ad7869165a9675844162b640c6917660aba150ee4ab60f0ff23c220f93336
semantic_inventory:
  - claim_key: CLAIM-3D430515651915F7
    claim_text: The read API must list, for a year, the yearly tax of each tax wrapper taxed on its balance, computed from the summed daily gross totals with cash of the wrapper's accounts only under a law in force, and otherwise saying it is not in force
    role: normative
    status: modeled
    span:
      start: 0
      end: 239
origin:
  kind: agent
  recorded_at: '2026-10-30T00:05:35.818Z'
id: REQ-views-yearly-taxes
type: req
---
The read API must list, for a year, the yearly tax of each tax wrapper taxed on its balance, computed from the summed daily gross totals with cash of the wrapper's accounts only under a law in force, and otherwise saying it is not in force.

## Context

Closed: superseded by REQ-views-oki-tax-days. D-127 computes OKI's tax from each account's days per exemption class, not from summed daily gross totals. Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.
