---
title: Safari's acceptance of the __Host- session cookie on http://localhost must be verified.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-85-2
  - lane:strict
text_ref: design-decisions:D-85#2
origin:
  kind: import
  ref: design-decisions:D-85#2
  recorded_at: '2026-10-29T08:23:39.964Z'
semantic_text: Safari's acceptance of the __Host- session cookie on http://localhost must be verified.
logic_claims:
  - CLAIM-45A5611D030A79FD
semantic_clauses:
  - Safari's acceptance of the __Host- session cookie on http://localhost must be verified.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: fb8472377a67213fd21e47e39a8cc0d6b6a7ae9d4d0550404de08eb2efb18db3
semantic_inventory:
  - claim_key: CLAIM-45A5611D030A79FD
    claim_text: Safari's acceptance of the __Host- session cookie on http://localhost must be verified
    role: normative
    status: modeled
    span:
      start: 0
      end: 86
id: REQ-ui-session-cookie-safari-localhost
type: req
---
Safari's acceptance of the __Host- session cookie on http://localhost must be verified.

## Context

D-85 keeps the __Host- prefix on the session cookie everywhere, localhost included (REQ-ui-session-cookie-host-prefix). A __Host- cookie must be Secure, and browsers differ on whether they accept a Secure cookie over plain http://localhost, so the author asks for Safari's behavior to be checked rather than assumed. The result decides whether the fallback in REQ-ui-session-cookie-localhost-fallback is ever used. The author recorded D-83 to D-85 after the design session that answered the implementation's questions.

## Source

> The session cookie uses the `__Host-` prefix everywhere, `localhost` included. Safari's acceptance of it on `http://localhost` is verified; only if Safari refuses it does `localhost`, and nothing else, fall back to a cookie without the prefix.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-85#2
