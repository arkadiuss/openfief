---
title: The core must reverse-proxy the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so that only the core ever calls a plugin.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-81-2
  - lane:strict
text_ref: design-decisions:D-81#2
origin:
  kind: import
  ref: design-decisions:D-81#2
  recorded_at: '2026-10-29T08:12:07.943Z'
semantic_text: The core must reverse-proxy the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so that only the core ever calls a plugin.
logic_claims:
  - CLAIM-791327F46736971D
semantic_clauses:
  - The core must reverse-proxy the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so that only the core ever calls a plugin.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 6fa5f15e4dd46e7bead50db6f22b7b4fd55d1793d31d5abf0a80bcb48252ace8
semantic_inventory:
  - claim_key: CLAIM-791327F46736971D
    claim_text: The core must reverse-proxy the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so that only the core ever calls a plugin
    role: normative
    status: modeled
    span:
      start: 0
      end: 159
id: REQ-core-reverse-proxies-view-plugins
type: req
---
The core must reverse-proxy the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so that only the core ever calls a plugin.

## Context

D-59 says only the core may call a plugin, with a NetworkPolicy admitting ingress from the core alone and a short-lived token for that plugin instance (REQ-plugins-core-only-caller); D-81 keeps that true for browser-facing view plugins by putting the core in front of them (REQ-core-browser-entry-point). Which paths the core forwards is declared in the plugin's manifest and approved in its grant (REQ-plugins-view-browser-path-grant). D-69 makes the built-in UI a backend-for-frontend (REQ-ui-backend-for-frontend). The author recorded D-81 and D-82 after the design session that answered the implementation's questions.

## Source

> The core is the only entry point for browsers. It reverse-proxies the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so only the core ever calls a plugin (D-59), and `kubectl port-forward` or the owner's own Ingress (D-28) needs to reach the core alone. A view plugin's manifest declares the path prefix it serves to browsers, and the grant approves it.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-81#2
