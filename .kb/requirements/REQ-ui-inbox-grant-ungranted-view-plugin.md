---
title: The Inbox offers the owner, on the item of a view plugin that can't read anything, one button, Grant the whole household, which makes the owner's household grant of that plugin instance through the core's CreateGrant, audited as any grant, and offers a read session none.
status: open
firmness: default
semantic_text: The Inbox offers the owner, on the item of a view plugin that can't read anything, one button, Grant the whole household, which makes the owner's household grant of that plugin instance through the core's CreateGrant, audited as any grant, and offers a read session none.
origin:
  kind: agent
  ref: 'design-session answers for the permissions plan, 2026-11-13 (later): attention item view-plugin-ungranted (D-196, D-209)'
  recorded_at: '2026-11-13T17:06:23.781Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 63b38597c23cb2534eb3475d873dca766713c1c63eef8203fe4ff4527d1cdd5c
semantic_inventory:
  - claim_key: CLAIM-10EA49BD5E426277
    claim_text: The Inbox offers the owner, on the item of a view plugin that can't read anything, one button, Grant the whole household, which makes the owner's household grant of that plugin instance through the core's CreateGrant, audited as any grant, and offers a read session none
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 270
logic_claims:
  - CLAIM-10EA49BD5E426277
id: REQ-ui-inbox-grant-ungranted-view-plugin
type: req
---
The Inbox offers the owner, on the item of a view plugin that can't read anything, one button, Grant the whole household, which makes the owner's household grant of that plugin instance through the core's CreateGrant, audited as any grant, and offers a read session none.

## Context

D-196 gives each item the actions that settle it, and the design session asked for one button on this item. The grant is the existing owner call, so the audit entry (D-211) and the SpiceDB write are the ones every grant has. The item also links to the View plugins section of the Access page, where the owner picks narrower grants.

## Source

> S4c: attention item `view-plugin-ungranted` (decisions -> missing settings, D-196): "<instance> can't read anything until you grant it accounts"; one action: grant the whole household, or open Access to pick. Raised for every installed view plugin with no D-209 grant while authz=spicedb; cleared when a grant exists; same sweep as model-route-broken; built-in UI exempt. CLI install summary for a view plugin: "grant it accounts in Settings -> Access".

Source: design-session answers for the permissions plan, 2026-11-13 (later).