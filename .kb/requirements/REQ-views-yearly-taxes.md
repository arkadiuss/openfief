---
title: The read API must list, for a year, the yearly tax of each tax wrapper taxed on its balance, computed from the summed daily gross totals with cash of the wrapper's accounts only under a law in force, and otherwise saying it is not in force.
status: open
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

D-76 computes OKI's yearly tax from stored daily valuations once the act is in force. Serving it in the read API and averaging valued days only are this package's readings. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-76]** Every account has a tax wrapper from the jurisdiction pack, which the owner sets when creating the account (D-72). The Polish pack's wrappers are a standard taxable account, IKE, IKZE and OKI (osobiste konto inwestycyjne). Inside OKI no Belka tax is due. Its yearly tax on the part of the average balance above the statutory limit, summed across all of the owner's OKI accounts, is computed from stored daily valuations only once the act is in force, with the limit and rate taken from the act.

docs/DESIGN.md#D-76
