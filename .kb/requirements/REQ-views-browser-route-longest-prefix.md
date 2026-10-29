---
title: The core must send a browser path it does not serve to the running view plugin whose granted browser path is the longest prefix covering it by whole segments, answer 404 when none covers it, and serve neither when two plugins claim that prefix.
status: open
text_ref: docs/DESIGN.md#D-81
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-81
  - lane:strict
semantic_text: The core must send a browser path it does not serve to the running view plugin whose granted browser path is the longest prefix covering it by whole segments, answer 404 when none covers it, and serve neither when two plugins claim that prefix.
logic_claims:
  - CLAIM-E7C8121653F304DA
semantic_clauses:
  - The core must send a browser path it does not serve to the running view plugin whose granted browser path is the longest prefix covering it by whole segments, answer 404 when none covers it, and serve neither when two plugins claim that prefix.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f073fc121071e8ccf069180f82b0f0e5517c8d29b9e687850be75434b034562c
semantic_inventory:
  - claim_key: CLAIM-E7C8121653F304DA
    claim_text: The core must send a browser path it does not serve to the running view plugin whose granted browser path is the longest prefix covering it by whole segments, answer 404 when none covers it, and serve neither when two plugins claim that prefix
    role: normative
    status: modeled
    span:
      start: 0
      end: 243
origin:
  kind: agent
  recorded_at: '2026-10-29T12:02:21.985Z'
id: REQ-views-browser-route-longest-prefix
type: req
---
The core must send a browser path it does not serve to the running view plugin whose granted browser path is the longest prefix covering it by whole segments, answer 404 when none covers it, and serve neither when two plugins claim that prefix.

## Context

D-81 has view plugins declare browser path prefixes that grants approve, but doesn't say how prefixes combine. Longest prefix is the usual reverse-proxy rule; serving neither of two equal claims is the narrowest reading, recorded as an open question in docs/implementation-questions.md.

## Source

> [D-81] The core is the only entry point for browsers. It reverse-proxies the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so only the core ever calls a plugin (D-59), and kubectl port-forward or the owner's own Ingress (D-28) needs to reach the core alone. A view plugin's manifest declares the path prefix it serves to browsers, and the grant approves it.

docs/DESIGN.md#D-81
