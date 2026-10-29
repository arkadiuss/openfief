---
title: A browser-facing view plugin's manifest must declare the path prefix it serves to browsers, and the grant must approve it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-81-3
  - lane:strict
text_ref: design-decisions:D-81#3
origin:
  kind: import
  ref: design-decisions:D-81#3
  recorded_at: '2026-10-29T08:12:52.335Z'
semantic_text: A browser-facing view plugin's manifest must declare the path prefix it serves to browsers, and the grant must approve it.
logic_claims:
  - CLAIM-51922FE81ABBCCEB
semantic_clauses:
  - A browser-facing view plugin's manifest must declare the path prefix it serves to browsers, and the grant must approve it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 9f273ea032cdcfdfa32242fe823a639e65f8a919a1170e7d07da0f407d595b5c
semantic_inventory:
  - claim_key: CLAIM-51922FE81ABBCCEB
    claim_text: A browser-facing view plugin's manifest must declare the path prefix it serves to browsers, and the grant must approve it
    role: normative
    status: modeled
    span:
      start: 0
      end: 121
id: REQ-plugins-view-browser-path-grant
type: req
---
A browser-facing view plugin's manifest must declare the path prefix it serves to browsers, and the grant must approve it.

## Context

D-81 has the core reverse-proxy a browser-facing view plugin's paths to it (REQ-core-reverse-proxies-view-plugins). D-16 keeps what a plugin's manifest requests apart from what the owner's grant approves, and the operator enforces only what is in both; D-42 makes the manifest a signed image label. REQ-ui-plugin-packaging records the built-in UI's manifest, which declares no browser path today. The author recorded D-81 and D-82 after the design session that answered the implementation's questions.

## Source

> The core is the only entry point for browsers. It reverse-proxies the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so only the core ever calls a plugin (D-59), and `kubectl port-forward` or the owner's own Ingress (D-28) needs to reach the core alone. A view plugin's manifest declares the path prefix it serves to browsers, and the grant approves it.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-81#3
