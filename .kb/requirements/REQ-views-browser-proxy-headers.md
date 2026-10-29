---
title: When the core proxies a browser request to a view plugin, it must present its own token for that instance and set the Host, X-Forwarded-Proto and X-Forwarded-Host itself, dropping every forwarding header the client sent.
status: open
text_ref: docs/DESIGN.md#D-81
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-81
  - lane:strict
semantic_text: When the core proxies a browser request to a view plugin, it must present its own token for that instance and set the Host, X-Forwarded-Proto and X-Forwarded-Host itself, dropping every forwarding header the client sent.
logic_claims:
  - CLAIM-14EE12207C3EB2E9
semantic_clauses:
  - When the core proxies a browser request to a view plugin, it must present its own token for that instance and set the Host, X-Forwarded-Proto and X-Forwarded-Host itself, dropping every forwarding header the client sent.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 69bfe320c127e569c2ef902b51214328ee743ba635a6ff0b9e33dd16aa40b731
semantic_inventory:
  - claim_key: CLAIM-14EE12207C3EB2E9
    claim_text: When the core proxies a browser request to a view plugin, it must present its own token for that instance and set the Host, X-Forwarded-Proto and X-Forwarded-Host itself, dropping every forwarding header the client sent
    role: condition
    status: modeled
    span:
      start: 0
      end: 219
origin:
  kind: agent
  recorded_at: '2026-10-29T12:02:12.887Z'
id: REQ-views-browser-proxy-headers
type: req
---
When the core proxies a browser request to a view plugin, it must present its own token for that instance and set the Host, X-Forwarded-Proto and X-Forwarded-Host itself, dropping every forwarding header the client sent.

## Context

D-81 makes the core the only entry point for browsers and D-59 has only the core call a plugin with its TokenRequest token. The UI builds the origin it hands the core's AuthService from Host and X-Forwarded-Proto, so after D-81 those must be what the core decided (orchestrator brief: trust only what the core sets). The scheme comes from TLS or the owner's Ingress; a forged one gains nothing, since origins are allow-listed.

## Source

> [D-81] The core is the only entry point for browsers. It reverse-proxies the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so only the core ever calls a plugin (D-59), and kubectl port-forward or the owner's own Ingress (D-28) needs to reach the core alone. A view plugin's manifest declares the path prefix it serves to browsers, and the grant approves it.

docs/DESIGN.md#D-81
