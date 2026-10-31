---
title: When two plugins claim a browser path at once, the one that doesn't keep it must stop, with the reason in its status.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-125-2
  - lane:strict
text_ref: design-decisions:D-125#2
origin:
  kind: import
  ref: design-decisions:D-125#2
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: When two plugins claim a browser path at once, the one that doesn't keep it must stop, with the reason in its status.
logic_claims:
  - CLAIM-344B9F3101E50D10
semantic_clauses:
  - When two plugins claim a browser path at once, the one that doesn't keep it must stop, with the reason in its status.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: d87dd03defb5b749a332b1a3f7cf7dff736cde9074fbf1b5b04f015b98e9f715
semantic_inventory:
  - claim_key: CLAIM-344B9F3101E50D10
    claim_text: When two plugins claim a browser path at once, the one that doesn't keep it must stop, with the reason in its status
    role: condition
    status: modeled
    span:
      start: 0
      end: 116
id: REQ-operator-contested-browser-path-stops
type: req
---
When two plugins claim a browser path at once, the one that doesn't keep it must stop, with the reason in its status.

## Context

D-114 has the operator not run a plugin with a browser path while another instance holds one, saying why in its status (REQ-operator-refuses-second-browser-view). When two such plugins are applied at once, the older keeps the path (REQ-operator-older-plugin-keeps-browser-path). The implementation raised the race as question #9 on docs/implementation-questions.md and tears the other down with Ready=False/BrowserPathHeld; on 2026-10-31 the author confirmed that the other stops with the reason in its status, as D-125.

## Source

> When two plugins claim a browser path at once, the older Plugin, by creation time and then by name, keeps it; the other stops, with the reason in its status.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-125#2
