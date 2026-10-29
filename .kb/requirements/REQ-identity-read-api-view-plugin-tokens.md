---
title: The read API must accept view plugins' own tokens.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-83-2
  - lane:strict
text_ref: design-decisions:D-83#2
origin:
  kind: import
  ref: design-decisions:D-83#2
  recorded_at: '2026-10-29T08:20:16.102Z'
semantic_text: The read API must accept view plugins' own tokens.
logic_claims:
  - CLAIM-00755B16EEA646D2
semantic_clauses:
  - The read API must accept view plugins' own tokens.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 1679b1993e4d80a53c1357aa00bd96c2bd82a0e86cc0f0d48b0b7913ca5827d0
semantic_inventory:
  - claim_key: CLAIM-00755B16EEA646D2
    claim_text: The read API must accept view plugins' own tokens
    role: normative
    status: modeled
    span:
      start: 0
      end: 49
id: REQ-identity-read-api-view-plugin-tokens
type: req
---
The read API must accept view plugins' own tokens.

## Context

D-6 has every view plugin, the built-in UI included, use the read API, and D-27 has plugins authenticate to the core with projected ServiceAccount tokens. D-83 settles which callers the read API accepts: sessions of either scope and script tokens (REQ-identity-read-api-sessions), view plugins' own tokens (this requirement), and never connector tokens (REQ-identity-read-api-refuses-connector-tokens). REQ-identity-read-api-plugins refines this clause with how the core checks such a token through the AuthService's caller check, and D-82 has the core check it with TokenReview (REQ-identity-view-plugin-token-review). The author recorded D-83 to D-85 after the design session that answered the implementation's questions.

## Source

> The read API accepts sessions of either scope, script tokens, and view plugins' own tokens. It refuses connector tokens: the core calls connectors (D-49), and a connector never needs to read the owner's record.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-83#2
