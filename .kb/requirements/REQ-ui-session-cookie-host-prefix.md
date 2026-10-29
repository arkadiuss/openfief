---
title: The session cookie must use the __Host- prefix everywhere, localhost included.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-85-1
  - lane:strict
text_ref: design-decisions:D-85#1
origin:
  kind: import
  ref: design-decisions:D-85#1
  recorded_at: '2026-10-29T08:23:01.797Z'
semantic_text: The session cookie must use the __Host- prefix everywhere, localhost included.
logic_claims:
  - CLAIM-1C67AFDE0EF8B830
semantic_clauses:
  - The session cookie must use the __Host- prefix everywhere, localhost included.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: fffc4298d7961749b9d14435582b0723fec75d8deecec054aac6b96c0779d61a
semantic_inventory:
  - claim_key: CLAIM-1C67AFDE0EF8B830
    claim_text: The session cookie must use the __Host- prefix everywhere, localhost included
    role: normative
    status: modeled
    span:
      start: 0
      end: 77
id: REQ-ui-session-cookie-host-prefix
type: req
---
The session cookie must use the __Host- prefix everywhere, localhost included.

## Context

D-69 makes the built-in UI a backend-for-frontend that keeps the owner's session for the browser, and D-60 has the owner create the first passkey on localhost, so the session cookie is also set on http://localhost. REQ-ui-session-cookie, which the author has not reviewed, already keeps the session in a __Host- cookie on HTTPS and on http://localhost alike. D-85 makes the prefix the author's rule everywhere; the only exception is REQ-ui-session-cookie-localhost-fallback, which applies only if Safari refuses the cookie on localhost (REQ-ui-session-cookie-safari-localhost). The author recorded D-83 to D-85 after the design session that answered the implementation's questions.

## Source

> The session cookie uses the `__Host-` prefix everywhere, `localhost` included. Safari's acceptance of it on `http://localhost` is verified; only if Safari refuses it does `localhost`, and nothing else, fall back to a cookie without the prefix.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-85#1
