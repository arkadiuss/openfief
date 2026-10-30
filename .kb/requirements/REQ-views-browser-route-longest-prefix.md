---
title: The core must send a browser path it does not serve to the one view plugin instance holding a granted browser path when that path covers it by whole segments, answer 404 when it does not or no instance holds one, and serve none while more than one instance holds a browser path.
status: open
text_ref: docs/DESIGN.md#D-114
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-114
  - lane:strict
semantic_text: The core must send a browser path it does not serve to the one view plugin instance holding a granted browser path when that path covers it by whole segments, answer 404 when it does not or no instance holds one, and serve none while more than one instance holds a browser path.
logic_claims:
  - CLAIM-3BEFBA00B96C998F
semantic_clauses:
  - The core must send a browser path it does not serve to the one view plugin instance holding a granted browser path when that path covers it by whole segments, answer 404 when it does not or no instance holds one, and serve none while more than one instance holds a browser path.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c11f7fae987bbbb52d07d100156b471051d3d4af6615756eca5fdd6329bc87a1
semantic_inventory:
  - claim_key: CLAIM-3BEFBA00B96C998F
    claim_text: The core must send a browser path it does not serve to the one view plugin instance holding a granted browser path when that path covers it by whole segments, answer 404 when it does not or no instance holds one, and serve none while more than one instance holds a browser path
    role: normative
    status: modeled
    span:
      start: 0
      end: 277
origin:
  kind: agent
  recorded_at: '2026-10-29T12:02:21.985Z'
id: REQ-views-browser-route-longest-prefix
type: req
---
The core must send a browser path it does not serve to the one view plugin instance holding a granted browser path when that path covers it by whole segments, answer 404 when it does not or no instance holds one, and serve none while more than one instance holds a browser path.

## Context

D-81 has view plugins declare browser path prefixes that grants approve; D-114 allows at most one instance to hold a browser path in M0, because the core passes the owner's session cookie to the view that serves a path. The CLI and the operator enforce it; the core's routing serves the one view and, as a last guard, none while it reads more than one, rather than pick one. This replaces the longest-prefix rule written before D-114, which let several views run on different prefixes.

## Source

> In M0, at most one plugin instance holds a browser path (D-81). The core passes the browser's cookies, the owner's session among them, to the view that serves a path, and the owner services accept any view plugin, so a second browser-facing view could act as the owner. The `openfief` CLI refuses to install a plugin that declares a browser path while another instance holds one, and the operator doesn't run such a plugin applied another way, saying why in its status.

docs/DESIGN.md#D-114
