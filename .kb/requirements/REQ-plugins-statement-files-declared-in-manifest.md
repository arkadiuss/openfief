---
title: A connector that takes statement files must say so in its signed manifest, as one sending cross-check prices does.
status: open
text_ref: design-decisions:D-139#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-139-1
  - lane:strict
semantic_text: A connector that takes statement files must say so in its signed manifest, as one sending cross-check prices does.
logic_claims:
  - CLAIM-BDB8F1A1FF387ACF
semantic_clauses:
  - A connector that takes statement files must say so in its signed manifest, as one sending cross-check prices does.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8da813ee92ddb99db80b364c5e42d8b08057f0827294cba6ea72eed32ef12f93
semantic_inventory:
  - claim_key: CLAIM-BDB8F1A1FF387ACF
    claim_text: A connector that takes statement files must say so in its signed manifest, as one sending cross-check prices does
    role: normative
    status: modeled
    span:
      start: 0
      end: 113
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-139#1
  recorded_at: '2026-11-01T12:23:25.569Z'
id: REQ-plugins-statement-files-declared-in-manifest
type: req
---
A connector that takes statement files must say so in its signed manifest, as one sending cross-check prices does.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #23. D-109 has the owner's UI show the upload form only on accounts whose connector takes files (REQ-ui-statement-upload-file-connectors-only), and D-112 has the manifest declare cross-check prices the same way. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A connector that takes statement files says so in its signed manifest, as one sending cross-check prices does (D-112). It is a capability the owner sees, not a permission the grant approves.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-139#1
