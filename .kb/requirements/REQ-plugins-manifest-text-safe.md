---
title: Manifest validation must refuse, with ErrUnsafeText, every manifest string, the config schema's keys and string values at any depth included, that holds a C0 or C1 control character, DEL, a Unicode bidirectional embedding, override or isolate, or invalid UTF-8.
status: open
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
semantic_text: Manifest validation must refuse, with ErrUnsafeText, every manifest string, the config schema's keys and string values at any depth included, that holds a C0 or C1 control character, DEL, a Unicode bidirectional embedding, override or isolate, or invalid UTF-8.
logic_claims:
  - CLAIM-6E4948DB14BF642A
semantic_clauses:
  - Manifest validation must refuse, with ErrUnsafeText, every manifest string, the config schema's keys and string values at any depth included, that holds a C0 or C1 control character, DEL, a Unicode bidirectional embedding, override or isolate, or invalid UTF-8.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4a8bc717ea077489cba57661bf6aa44b7793a283ce142445c51e29361ba2ec29
semantic_inventory:
  - claim_key: CLAIM-6E4948DB14BF642A
    claim_text: Manifest validation must refuse, with ErrUnsafeText, every manifest string, the config schema's keys and string values at any depth included, that holds a C0 or C1 control character, DEL, a Unicode bidirectional embedding, override or isolate, or invalid UTF-8
    role: normative
    status: modeled
    span:
      start: 0
      end: 260
origin:
  kind: agent
  recorded_at: '2026-11-03T09:18:18.358Z'
id: REQ-plugins-manifest-text-safe
type: req
---
Manifest validation must refuse, with ErrUnsafeText, every manifest string, the config schema's keys and string values at any depth included, that holds a C0 or C1 control character, DEL, a Unicode bidirectional embedding, override or isolate, or invalid UTF-8.

## Context

Recorded with default firmness by the implementation agent of the second security review's fixes (fix/review2-r1-r4), from finding R4: a signed plugin's config schema property named "\x1b[6A\r\x1b[2K  Egress hosts: none" moved the cursor up and overwrote the consent screen's egress line, because validation refused only CR and LF in one field and the CLI printed descriptions, property names and enum values raw. The owner confirms what the screen shows, so the screen must show what the manifest says.

## Source

> [D-46] Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` ...

> [D-90] ... `openfief plugin upgrade <instance> <version>` verifies the new image, shows how its requested permissions differ, and applies the upgrade only when the owner confirms.

docs/DESIGN.md#D-46, docs/DESIGN.md#D-90
