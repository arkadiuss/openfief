---
title: The operator must enforce only permissions that are both requested and approved.
status: open
text_ref: design-decisions:D-16#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-16-2
  - lane:strict
semantic_text: The operator must enforce only permissions that are both requested and approved.
logic_claims:
  - CLAIM-5D6215D6D4158888
semantic_clauses:
  - The operator must enforce only permissions that are both requested and approved.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c6882ebb59cf656ba42ba90561602d53ed438f585d81cec5695360df77355be8
semantic_inventory:
  - claim_key: CLAIM-5D6215D6D4158888
    claim_text: The operator must enforce only permissions that are both requested and approved
    role: normative
    status: modeled
    span:
      start: 0
      end: 79
origin:
  kind: import
  ref: design-decisions:D-16#2
  recorded_at: '2026-10-27T20:12:30.303Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-12177128A788CEDC
type: req
---
The operator must enforce only permissions that are both requested and approved.

## Source

> Plugin permissions are split into the `Plugin` resource (what the vendor requests) and a `PluginGrant` (what the user approved). The operator enforces only what is in both. An upgrade that requests more waits for approval.

Source: OpenFief author decisions (DESIGN.md section 10) - D-16#2
