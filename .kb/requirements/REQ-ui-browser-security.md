---
title: The built-in UI must refuse every state-changing request that a browser did not send from the UI's own origin, and serve every response with a Content-Security-Policy that allows scripts and styles from the UI itself only, no inline code and no framing.
status: open
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
semantic_text: The built-in UI must refuse every state-changing request that a browser did not send from the UI's own origin, and serve every response with a Content-Security-Policy that allows scripts and styles from the UI itself only, no inline code and no framing.
logic_claims:
  - CLAIM-0318328C509EFF88
semantic_clauses:
  - The built-in UI must refuse every state-changing request that a browser did not send from the UI's own origin, and serve every response with a Content-Security-Policy that allows scripts and styles from the UI itself only, no inline code and no framing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 89740f0a6f0a705b36c69aeda0c320dd4ee49e7f20eab62e6556f5d528f68186
semantic_inventory:
  - claim_key: CLAIM-0318328C509EFF88
    claim_text: The built-in UI must refuse every state-changing request that a browser did not send from the UI's own origin, and serve every response with a Content-Security-Policy that allows scripts and styles from the UI itself only, no inline code and no framing
    role: normative
    status: modeled
    span:
      start: 0
      end: 252
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:46.042Z'
id: REQ-ui-browser-security
type: req
---
The built-in UI must refuse every state-changing request that a browser did not send from the UI's own origin, and serve every response with a Content-Security-Policy that allows scripts and styles from the UI itself only, no inline code and no framing.

## Context

The orchestrator's task for the UI: strict CSP (self or nonce only, frame-ancestors none), security headers, CSRF defense by SameSite=Strict plus an Origin/Sec-Fetch-Site check, htmx and the chart library vendored, pinned and served by the UI, no CDN.

## Source

> Strict CSP: no inline scripts or styles, nonce or self only, frame-ancestors none.

orchestrator task for plugins/ui
