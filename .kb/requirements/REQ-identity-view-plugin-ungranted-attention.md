---
title: While the core runs with SpiceDB authorization, every installed view plugin that is Ready, holds the view capability, serves no browser path and holds no grant in force raises one attention item among the Inbox's decisions, naming its instance and saying it can't read anything until the owner grants it accounts, which clears when a grant exists or the plugin is uninstalled; none is raised under the policy, nor for the built-in UI, which reads with the signed-in person's session.
status: open
firmness: default
semantic_text: While the core runs with SpiceDB authorization, every installed view plugin that is Ready, holds the view capability, serves no browser path and holds no grant in force raises one attention item among the Inbox's decisions, naming its instance and saying it can't read anything until the owner grants it accounts, which clears when a grant exists or the plugin is uninstalled; none is raised under the policy, nor for the built-in UI, which reads with the signed-in person's session.
origin:
  kind: agent
  ref: 'design-session answers for the permissions plan, 2026-11-13 (later): attention item view-plugin-ungranted (D-196, D-209)'
  recorded_at: '2026-11-13T17:06:22.049Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8e301e5bb43d6a369db8c5aa657a2202c108f89657f09834e5e81555ae555f51
semantic_inventory:
  - claim_key: CLAIM-9A74175455437963
    claim_text: While the core runs with SpiceDB authorization, every installed view plugin that is Ready, holds the view capability, serves no browser path and holds no grant in force raises one attention item among the Inbox's decisions, naming its instance and saying it can't read anything until the owner grants it accounts, which clears when a grant exists or the plugin is uninstalled
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 375
  - claim_key: CLAIM-C49700DF697C34CD
    claim_text: none is raised under the policy, nor for the built-in UI, which reads with the signed-in person's session
    role: descriptive
    status: missing
    span:
      start: 377
      end: 482
logic_claims:
  - CLAIM-9A74175455437963
  - CLAIM-C49700DF697C34CD
id: REQ-identity-view-plugin-ungranted-attention
type: req
---
While the core runs with SpiceDB authorization, every installed view plugin that is Ready, holds the view capability, serves no browser path and holds no grant in force raises one attention item among the Inbox's decisions, naming its instance and saying it can't read anything until the owner grants it accounts, which clears when a grant exists or the plugin is uninstalled; none is raised under the policy, nor for the built-in UI, which reads with the signed-in person's session.

## Context

D-209's cutover gives the household grant only to the view plugins that hold read access when SpiceDB takes over, so a view plugin installed later, such as Ghostfolio, starts with none and reads nothing. The design session after the cutover asked that the owner be told in the Inbox, by the same periodic sweep as the broken model routes, and that the built-in UI, which reads with the signed-in person's session, never its plugin identity, be exempt. Under the policy every holder of read access views every account, so there is nothing to grant.

## Source

> S4c: attention item `view-plugin-ungranted` (decisions -> missing settings, D-196): "<instance> can't read anything until you grant it accounts"; one action: grant the whole household, or open Access to pick. Raised for every installed view plugin with no D-209 grant while authz=spicedb; cleared when a grant exists; same sweep as model-route-broken; built-in UI exempt. CLI install summary for a view plugin: "grant it accounts in Settings -> Access".

Source: design-session answers for the permissions plan, 2026-11-13 (later)