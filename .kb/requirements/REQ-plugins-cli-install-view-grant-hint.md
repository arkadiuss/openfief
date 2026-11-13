---
title: The install of a view plugin ends its summary with a line telling the owner that, if OpenFief runs with SpiceDB authorization, they grant it accounts in Settings → Access, and the install of a plugin without the view capability prints none.
status: open
firmness: default
semantic_text: The install of a view plugin ends its summary with a line telling the owner that, if OpenFief runs with SpiceDB authorization, they grant it accounts in Settings → Access, and the install of a plugin without the view capability prints none.
origin:
  kind: agent
  ref: 'design-session answers for the permissions plan, 2026-11-13 (later): attention item view-plugin-ungranted (D-196, D-209)'
  recorded_at: '2026-11-13T17:06:25.781Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f944c932b057e92eb15dfd3042cd4a7d9664231db125e8cae3dabdb8ba6d0ec5
semantic_inventory:
  - claim_key: CLAIM-D059509CB531BE28
    claim_text: The install of a view plugin ends its summary with a line telling the owner that, if OpenFief runs with SpiceDB authorization, they grant it accounts in Settings → Access, and the install of a plugin without the view capability prints none
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 241
logic_claims:
  - CLAIM-D059509CB531BE28
id: REQ-plugins-cli-install-view-grant-hint
type: req
---
The install of a view plugin ends its summary with a line telling the owner that, if OpenFief runs with SpiceDB authorization, they grant it accounts in Settings → Access, and the install of a plugin without the view capability prints none.

## Context

A view plugin reads nothing under SpiceDB authorization until the owner grants it accounts. The command line can't tell which authorization the core runs with, so the design session asked for a line that is correct either way.

## Source

> S4c: attention item `view-plugin-ungranted` (decisions -> missing settings, D-196): "<instance> can't read anything until you grant it accounts"; one action: grant the whole household, or open Access to pick. Raised for every installed view plugin with no D-209 grant while authz=spicedb; cleared when a grant exists; same sweep as model-route-broken; built-in UI exempt. CLI install summary for a view plugin: "grant it accounts in Settings -> Access".

Source: design-session answers for the permissions plan, 2026-11-13 (later)