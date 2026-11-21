---
title: The install of a view plugin ends its summary with a line telling the owner that it reads nothing until they grant it accounts in Settings → Access, whichever authorization OpenFief runs, and the install of a plugin without the view capability prints none.
status: open
firmness: default
semantic_text: The install of a view plugin ends its summary with a line telling the owner that it reads nothing until they grant it accounts in Settings → Access, whichever authorization OpenFief runs, and the install of a plugin without the view capability prints none.
origin:
  kind: agent
  ref: 'design-session answers for the permissions plan, 2026-11-13 (later): attention item view-plugin-ungranted (D-196, D-209)'
  recorded_at: '2026-11-13T17:06:25.781Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7ddcacc7903297cb2a288faf5f6a3eea5293d93ee5103e3e2881465d3e1ff408
semantic_inventory:
  - claim_key: CLAIM-A751090D7E937866
    claim_text: The install of a view plugin ends its summary with a line telling the owner that it reads nothing until they grant it accounts in Settings → Access, whichever authorization OpenFief runs, and the install of a plugin without the view capability prints none
    role: descriptive
    status: missing
    span:
      start: 0
      end: 257
logic_claims:
  - CLAIM-A751090D7E937866
id: REQ-plugins-cli-install-view-grant-hint
type: req
---
The install of a view plugin ends its summary with a line telling the owner that it reads nothing until they grant it accounts in Settings → Access, whichever authorization OpenFief runs, and the install of a plugin without the view capability prints none.

## Context

A view plugin reads nothing until the owner grants it accounts, under the policy and under SpiceDB alike, so the line is unconditional. The first wording said "if OpenFief runs with SpiceDB authorization", because the command line can't tell which authorization the core runs with; since the policy enforces grants too, it no longer needs to. Changed on 2026-11-21 at the coordinator's request.

## Source

> S4c: attention item `view-plugin-ungranted` (decisions -> missing settings, D-196): "<instance> can't read anything until you grant it accounts"; one action: grant the whole household, or open Access to pick. Raised for every installed view plugin with no D-209 grant while authz=spicedb; cleared when a grant exists; same sweep as model-route-broken; built-in UI exempt. CLI install summary for a view plugin: "grant it accounts in Settings -> Access".

Source: design-session answers for the permissions plan, 2026-11-13 (later)
