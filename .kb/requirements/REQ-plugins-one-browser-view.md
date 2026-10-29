---
title: In M0, at most one plugin instance must hold a browser path.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-114-1
  - lane:strict
text_ref: design-decisions:D-114#1
origin:
  kind: import
  ref: design-decisions:D-114#1
  recorded_at: '2026-10-29T16:10:51.000Z'
semantic_text: In M0, at most one plugin instance must hold a browser path.
logic_claims:
  - CLAIM-0E853DC78E08CA92
semantic_clauses:
  - In M0, at most one plugin instance must hold a browser path.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 3b992eb58873f63bc0da2734afb9ae210e491fb3928756ec34c11bd831eb2a19
semantic_inventory:
  - claim_key: CLAIM-0E853DC78E08CA92
    claim_text: In M0, at most one plugin instance must hold a browser path
    role: normative
    status: modeled
    span:
      start: 0
      end: 59
id: REQ-plugins-one-browser-view
type: req
---
In M0, at most one plugin instance must hold a browser path.

## Context

The core passes the browser's cookies, the owner's session among them, to the view plugin that serves a browser path (D-81, REQ-core-reverse-proxies-view-plugins), and the owner services accept any view plugin (D-79, REQ-core-owner-entry-service-callers), so a second browser-facing view could act as the owner. The author decided D-114 while answering the implementation's question about two view plugins with the same browser path, which until then both got 503 (REQ-views-browser-route-longest-prefix). REQ-plugins-cli-refuses-second-browser-view and REQ-operator-refuses-second-browser-view record how the CLI and the operator keep to it.

## Source

> In M0, at most one plugin instance holds a browser path (D-81). The core passes the browser's cookies, the owner's session among them, to the view that serves a path, and the owner services accept any view plugin, so a second browser-facing view could act as the owner. The `openfief` CLI refuses to install a plugin that declares a browser path while another instance holds one, and the operator doesn't run such a plugin applied another way, saying why in its status.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-114#1
