---
title: The openfief CLI must refuse to install a plugin that declares a browser path while another instance holds one.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-114-2
  - lane:strict
text_ref: design-decisions:D-114#2
origin:
  kind: import
  ref: design-decisions:D-114#2
  recorded_at: '2026-10-29T16:10:51.000Z'
semantic_text: The openfief CLI must refuse to install a plugin that declares a browser path while another instance holds one.
logic_claims:
  - CLAIM-EDB65D6B77D16386
semantic_clauses:
  - The openfief CLI must refuse to install a plugin that declares a browser path while another instance holds one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2f45a196c28e059f0608e06cfd815fbeba8ab5701adcad2bb8908a4b46b99ae4
semantic_inventory:
  - claim_key: CLAIM-EDB65D6B77D16386
    claim_text: The openfief CLI must refuse to install a plugin that declares a browser path while another instance holds one
    role: normative
    status: modeled
    span:
      start: 0
      end: 110
id: REQ-plugins-cli-refuses-second-browser-view
type: req
---
The openfief CLI must refuse to install a plugin that declares a browser path while another instance holds one.

## Context

In M0 at most one plugin instance holds a browser path (REQ-plugins-one-browser-view): the core passes the browser's cookies, the owner's session among them, to the view that serves a path, and the owner services accept any view plugin, so a second browser-facing view could act as the owner. The CLI is where the owner installs plugins (D-46, REQ-plugins-cli-install), so it refuses such an install before applying anything; REQ-operator-refuses-second-browser-view covers a Plugin applied another way. The author decided D-114 while answering the implementation's question about two view plugins with the same browser path, which until then both got 503.

## Source

> In M0, at most one plugin instance holds a browser path (D-81). The core passes the browser's cookies, the owner's session among them, to the view that serves a path, and the owner services accept any view plugin, so a second browser-facing view could act as the owner. The `openfief` CLI refuses to install a plugin that declares a browser path while another instance holds one, and the operator doesn't run such a plugin applied another way, saying why in its status.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-114#2
