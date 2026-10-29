---
title: A session cookie without the __Host- prefix must be used only on localhost, and only if Safari refuses the prefixed cookie there.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-85-3
  - lane:strict
text_ref: design-decisions:D-85#3
origin:
  kind: import
  ref: design-decisions:D-85#3
  recorded_at: '2026-10-29T08:24:16.111Z'
semantic_text: A session cookie without the __Host- prefix must be used only on localhost, and only if Safari refuses the prefixed cookie there.
logic_claims:
  - CLAIM-FA40DAA3AC4F2570
semantic_clauses:
  - A session cookie without the __Host- prefix must be used only on localhost, and only if Safari refuses the prefixed cookie there.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 5c8bdd7e675b5706d77da59e3c4520f8c9700d1e07c47965c727e8765804707e
semantic_inventory:
  - claim_key: CLAIM-FA40DAA3AC4F2570
    claim_text: A session cookie without the __Host- prefix must be used only on localhost, and only if Safari refuses the prefixed cookie there
    role: normative
    status: modeled
    span:
      start: 0
      end: 128
id: REQ-ui-session-cookie-localhost-fallback
type: req
---
A session cookie without the __Host- prefix must be used only on localhost, and only if Safari refuses the prefixed cookie there.

## Context

D-85 keeps the __Host- prefix on the session cookie everywhere, localhost included (REQ-ui-session-cookie-host-prefix), and has Safari's acceptance of it on http://localhost verified (REQ-ui-session-cookie-safari-localhost). This requirement bounds the one exception the author allows: if Safari refuses the prefixed cookie on http://localhost, localhost and no other host falls back to a cookie without the prefix. The author recorded D-83 to D-85 after the design session that answered the implementation's questions.

## Source

> The session cookie uses the `__Host-` prefix everywhere, `localhost` included. Safari's acceptance of it on `http://localhost` is verified; only if Safari refuses it does `localhost`, and nothing else, fall back to a cookie without the prefix.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-85#3
