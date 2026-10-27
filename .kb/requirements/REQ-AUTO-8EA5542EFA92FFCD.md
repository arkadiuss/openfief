---
title: A plugin upgrade that requests more permissions must not deploy until they are approved.
status: open
text_ref: design-decisions:D-16#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-16-3
  - lane:strict
semantic_text: A plugin upgrade that requests more permissions must not deploy until they are approved.
logic_claims:
  - CLAIM-4E50A2E1D2C2CAD3
semantic_clauses:
  - A plugin upgrade that requests more permissions must not deploy until they are approved.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9e31a7461ef8cfe6a1227e8925f0280f61ab28edc7a46ea0ba006dd97ed11d6f
semantic_inventory:
  - claim_key: CLAIM-4E50A2E1D2C2CAD3
    claim_text: A plugin upgrade that requests more permissions must not deploy until they are approved
    role: normative
    status: modeled
    span:
      start: 0
      end: 87
origin:
  kind: import
  ref: design-decisions:D-16#3
  recorded_at: '2026-10-27T20:12:30.656Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-8EA5542EFA92FFCD
type: req
---
A plugin upgrade that requests more permissions must not deploy until they are approved.

## Source

> Plugin permissions are split into the `Plugin` resource (what the vendor requests) and a `PluginGrant` (what the user approved). The operator enforces only what is in both. An upgrade that requests more waits for approval.

Source: OpenFief author decisions (DESIGN.md section 10) - D-16#3
