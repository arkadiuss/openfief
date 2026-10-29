---
title: The operator must not run a plugin that declares a browser path while another instance holds one, and must say why in its status.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-114-3
  - lane:strict
text_ref: design-decisions:D-114#3
origin:
  kind: import
  ref: design-decisions:D-114#3
  recorded_at: '2026-10-29T16:10:51.000Z'
semantic_text: The operator must not run a plugin that declares a browser path while another instance holds one, and must say why in its status.
logic_claims:
  - CLAIM-60931CA73EB2C477
semantic_clauses:
  - The operator must not run a plugin that declares a browser path while another instance holds one, and must say why in its status.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c51eb3b6bfde6675e623e1252307b82df90ad92ff2099ba2a2135c3a4d64531c
semantic_inventory:
  - claim_key: CLAIM-60931CA73EB2C477
    claim_text: The operator must not run a plugin that declares a browser path while another instance holds one, and must say why in its status
    role: normative
    status: modeled
    span:
      start: 0
      end: 128
id: REQ-operator-refuses-second-browser-view
type: req
---
The operator must not run a plugin that declares a browser path while another instance holds one, and must say why in its status.

## Context

In M0 at most one plugin instance holds a browser path (REQ-plugins-one-browser-view): the core passes the browser's cookies, the owner's session among them, to the view that serves a path, and the owner services accept any view plugin, so a second browser-facing view could act as the owner. The CLI refuses such an install (REQ-plugins-cli-refuses-second-browser-view), but a Plugin can also be applied another way, such as with kubectl or GitOps, so the operator refuses to run it and records the reason in the Plugin's status. The author decided D-114 while answering the implementation's question about two view plugins with the same browser path, which until then both got 503.

## Source

> In M0, at most one plugin instance holds a browser path (D-81). The core passes the browser's cookies, the owner's session among them, to the view that serves a path, and the owner services accept any view plugin, so a second browser-facing view could act as the owner. The `openfief` CLI refuses to install a plugin that declares a browser path while another instance holds one, and the operator doesn't run such a plugin applied another way, saying why in its status.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-114#3
