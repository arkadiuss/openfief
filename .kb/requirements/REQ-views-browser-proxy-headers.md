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

D-81 makes the core the only entry point for browsers, which reach its browser port through the owner's route (D-28) or a port-forward; D-59 has the core present its own token to every plugin. A view must trust what the core sets and nothing a client forged.

## Source

> The core is the only entry point for browsers. It reverse-proxies the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so only the core ever calls a plugin (D-59).

docs/DESIGN.md#D-81
