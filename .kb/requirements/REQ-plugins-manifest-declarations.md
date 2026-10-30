---
title: A plugin manifest may declare crossCheckPrices and statementFiles as booleans that are signed with the image and carried in the Plugin resource like every other field, and the core must read them from the running version's manifest.
status: open
text_ref: docs/DESIGN.md#D-112
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-112
  - lane:strict
semantic_text: A plugin manifest may declare crossCheckPrices and statementFiles as booleans that are signed with the image and carried in the Plugin resource like every other field, and the core must read them from the running version's manifest.
logic_claims:
  - CLAIM-C887862CD7D3E135
semantic_clauses:
  - A plugin manifest may declare crossCheckPrices and statementFiles as booleans that are signed with the image and carried in the Plugin resource like every other field, and the core must read them from the running version's manifest.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dd8a742fe1bf3489e01b3af4f7d40dfc322aa7b09f40e6b6d6ad5302cb81695a
semantic_inventory:
  - claim_key: CLAIM-C887862CD7D3E135
    claim_text: A plugin manifest may declare crossCheckPrices and statementFiles as booleans that are signed with the image and carried in the Plugin resource like every other field, and the core must read them from the running version's manifest
    role: normative
    status: modeled
    span:
      start: 0
      end: 231
origin:
  kind: agent
  recorded_at: '2026-10-30T04:11:05.502Z'
id: REQ-plugins-manifest-declarations
type: req
---
A plugin manifest may declare crossCheckPrices and statementFiles as booleans that are signed with the image and carried in the Plugin resource like every other field, and the core must read them from the running version's manifest.

## Context

D-112 makes the manifest declare who may send cross-check prices; D-109 needs the core to know which connectors take files, which no decision places, so the same manifest declaration is the narrowest reading (docs/implementation-questions.md).

## Source

> Only a connector whose manifest declares it may send cross-check prices. The owner can acknowledge a flagged check, which is recorded append-only, and the UI shows the latest check per instrument with its history.

docs/DESIGN.md#D-112
