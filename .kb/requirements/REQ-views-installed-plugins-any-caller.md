---
title: The list of installed plugins, with their image digests and signers, must be readable by any read API caller, the demo's read-only visitors included.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-86-1
  - lane:strict
text_ref: design-decisions:D-86#1
origin:
  kind: import
  ref: design-decisions:D-86#1
  recorded_at: '2026-10-29T08:26:47.591Z'
semantic_text: The list of installed plugins, with their image digests and signers, must be readable by any read API caller, the demo's read-only visitors included.
logic_claims:
  - CLAIM-AFCA5A2DA73F39AE
semantic_clauses:
  - The list of installed plugins, with their image digests and signers, must be readable by any read API caller, the demo's read-only visitors included.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: abe9f4e641fef01e5e1d19611ed5f17f579abddc7da28667518f2516b5f1b82b
semantic_inventory:
  - claim_key: CLAIM-AFCA5A2DA73F39AE
    claim_text: The list of installed plugins, with their image digests and signers, must be readable by any read API caller, the demo's read-only visitors included
    role: normative
    status: modeled
    span:
      start: 0
      end: 148
id: REQ-views-installed-plugins-any-caller
type: req
---
The list of installed plugins, with their image digests and signers, must be readable by any read API caller, the demo's read-only visitors included.

## Context

D-46 has the UI only show installed plugins in M0, and D-83 settles who calls the read API: sessions of either scope, script tokens and view plugins' own tokens. D-86 makes the installed-plugin list readable by all of them, the demo's read-only visitors included, so that anyone using OpenFief can see what runs in it and who signed it. REQ-views-installed-plugins and REQ-ui-installed-plugins, which the author has not reviewed, refine how the read API and the UI present that list. The author recorded D-86 and D-87 after the design session that answered the implementation's questions.

## Source

> The list of installed plugins, with their image digests and signers, is readable by any read API caller (D-83), including the demo's read-only visitors: anyone using OpenFief can see what runs in it and who signed it.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-86#1
