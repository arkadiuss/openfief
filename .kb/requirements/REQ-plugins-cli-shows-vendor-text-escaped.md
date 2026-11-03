---
title: The CLI must print every string a plugin's vendor, its catalog entry or the cluster wrote with Go's visible escapes whenever it holds a character that doesn't print as itself, and must escape any control or bidirectional character left in what it writes to the terminal.
status: open
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
semantic_text: The CLI must print every string a plugin's vendor, its catalog entry or the cluster wrote with Go's visible escapes whenever it holds a character that doesn't print as itself, and must escape any control or bidirectional character left in what it writes to the terminal.
logic_claims:
  - CLAIM-EB5DB86C6F27DFCC
semantic_clauses:
  - The CLI must print every string a plugin's vendor, its catalog entry or the cluster wrote with Go's visible escapes whenever it holds a character that doesn't print as itself, and must escape any control or bidirectional character left in what it writes to the terminal.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3f50401ad0e2c8f7e281b316b240f17cd83b6f6608068cda67a071125eacec24
semantic_inventory:
  - claim_key: CLAIM-EB5DB86C6F27DFCC
    claim_text: The CLI must print every string a plugin's vendor, its catalog entry or the cluster wrote with Go's visible escapes whenever it holds a character that doesn't print as itself, and must escape any control or bidirectional character left in what it writes to the terminal
    role: normative
    status: modeled
    span:
      start: 0
      end: 269
origin:
  kind: agent
  recorded_at: '2026-11-03T09:18:27.801Z'
id: REQ-plugins-cli-shows-vendor-text-escaped
type: req
---
The CLI must print every string a plugin's vendor, its catalog entry or the cluster wrote with Go's visible escapes whenever it holds a character that doesn't print as itself, and must escape any control or bidirectional character left in what it writes to the terminal.

## Context

Recorded with default firmness by the implementation agent of the second security review's fixes (fix/review2-r1-r4), from finding R4: a signed plugin's config schema property named "\x1b[6A\r\x1b[2K  Egress hosts: none" moved the cursor up and overwrote the consent screen's egress line, because validation refused only CR and LF in one field and the CLI printed descriptions, property names and enum values raw. The owner confirms what the screen shows, so the screen must show what the manifest says.

## Source

> [D-46] Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` ...

> [D-90] ... `openfief plugin upgrade <instance> <version>` verifies the new image, shows how its requested permissions differ, and applies the upgrade only when the owner confirms.

docs/DESIGN.md#D-46, docs/DESIGN.md#D-90
