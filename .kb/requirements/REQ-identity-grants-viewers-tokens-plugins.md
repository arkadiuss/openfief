---
title: Viewers, script tokens and view plugins are all granted this way.
status: open
text_ref: design-decisions:D-209#4
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-209-4
semantic_text: Viewers, script tokens and view plugins are all granted this way.
logic_claims:
  - CLAIM-6A3DA11B86E984FA
semantic_clauses:
  - Viewers, script tokens and view plugins are all granted this way.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dde22772fae4bf108cfd363cab8628ef9a2d5a8deb58073c60d9c269cab12e25
semantic_inventory:
  - claim_key: CLAIM-6A3DA11B86E984FA
    claim_text: Viewers, script tokens and view plugins are all granted this way
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 64
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-209#4
  recorded_at: '2026-11-13T10:00:37.862Z'
id: REQ-identity-grants-viewers-tokens-plugins
type: req
---
Viewers, script tokens and view plugins are all granted this way.

## Context

Part of D-209. Script tokens and view plugins are granted today; viewers (D-65) come with the second alpha, and the grant model already names them as a kind of holder, which CreateGrant refuses until they exist. A view plugin's grant is separate from its PluginGrant, which approves what the plugin may do in the cluster, and it ends when the plugin instance is uninstalled. The decision states no further reason. The author recorded D-209 in DESIGN.md.

## Source

> The owner grants access per account, per sleeve or for the whole household; a sleeve's grant covers the accounts added to it later, and the household's every account, present and future. Viewers, script tokens and view plugins are all granted this way, and a grant only ever lets its holder view. When SpiceDB takes over (D-164), every script token and view plugin then holding read access is given the household grant, shown on the Access page, so none loses or gains access silently.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-209#4