---
title: The read API must refuse connector tokens.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-83-3
  - lane:strict
text_ref: design-decisions:D-83#3
origin:
  kind: import
  ref: design-decisions:D-83#3
  recorded_at: '2026-10-29T08:20:51.377Z'
semantic_text: The read API must refuse connector tokens.
logic_claims:
  - CLAIM-62CDD8F02D322AAC
semantic_clauses:
  - The read API must refuse connector tokens.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 57b57002cd7aa48a17ebbcf44040ad20a74e382f8062a6ca8655d77f745a41f3
semantic_inventory:
  - claim_key: CLAIM-62CDD8F02D322AAC
    claim_text: The read API must refuse connector tokens
    role: normative
    status: modeled
    span:
      start: 0
      end: 41
id: REQ-identity-read-api-refuses-connector-tokens
type: req
---
The read API must refuse connector tokens.

## Context

D-83 settles which callers the read API accepts: sessions of either scope and script tokens (REQ-identity-read-api-sessions) and view plugins' own tokens (REQ-identity-read-api-view-plugin-tokens). It refuses connector tokens because the core calls connectors (D-49), not the other way round, and a connector never needs to read the owner's record. The AuthService's caller check already accepts only view plugins (D-69, REQ-identity-auth-callers). The author recorded D-83 to D-85 after the design session that answered the implementation's questions.

## Source

> The read API accepts sessions of either scope, script tokens, and view plugins' own tokens. It refuses connector tokens: the core calls connectors (D-49), and a connector never needs to read the owner's record.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-83#3
