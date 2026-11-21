---
title: Under either authorization, the policy or SpiceDB, every installed view plugin that is Ready, holds the view capability, serves no browser path and holds no grant in force raises one attention item among the Inbox's decisions, naming its instance and saying it can't read anything until the owner grants it accounts, which clears when a grant exists or the plugin is uninstalled; none is raised for the built-in UI, which reads with the signed-in person's session.
status: open
firmness: default
semantic_text: Under either authorization, the policy or SpiceDB, every installed view plugin that is Ready, holds the view capability, serves no browser path and holds no grant in force raises one attention item among the Inbox's decisions, naming its instance and saying it can't read anything until the owner grants it accounts, which clears when a grant exists or the plugin is uninstalled; none is raised for the built-in UI, which reads with the signed-in person's session.
origin:
  kind: agent
  ref: 'design-session answers for the permissions plan, 2026-11-13 (later): attention item view-plugin-ungranted (D-196, D-209)'
  recorded_at: '2026-11-13T17:06:22.049Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: af3d6369077eab9d9be957445e6db9c3b9087a99d7a26882d1a80f18fa46d15d
semantic_inventory:
  - claim_key: CLAIM-67D54EC5E5AC463C
    claim_text: Under either authorization, the policy or SpiceDB, every installed view plugin that is Ready, holds the view capability, serves no browser path and holds no grant in force raises one attention item among the Inbox's decisions, naming its instance and saying it can't read anything until the owner grants it accounts, which clears when a grant exists or the plugin is uninstalled
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 378
    subjects:
      - view_plugin.ungranted_attention_item
  - claim_key: CLAIM-FDCB0209280E7441
    claim_text: none is raised for the built-in UI, which reads with the signed-in person's session
    role: descriptive
    status: missing
    span:
      start: 380
      end: 463
logic_claims:
  - CLAIM-67D54EC5E5AC463C
  - CLAIM-FDCB0209280E7441
id: REQ-identity-view-plugin-ungranted-attention
type: req
---
Under either authorization, the policy or SpiceDB, every installed view plugin that is Ready, holds the view capability, serves no browser path and holds no grant in force raises one attention item among the Inbox's decisions, naming its instance and saying it can't read anything until the owner grants it accounts, which clears when a grant exists or the plugin is uninstalled; none is raised for the built-in UI, which reads with the signed-in person's session.

## Context

D-209's cutover gives the household grant only to the view plugins that hold read access when the grants begin to limit them, so a view plugin installed later, such as Ghostfolio, starts with none and reads nothing. The design session after the cutover asked that the owner be told in the Inbox, by the same periodic sweep as the broken model routes, and that the built-in UI, which reads with the signed-in person's session, never its plugin identity, be exempt. The session wrote it for SpiceDB only, on the premise that under the policy every holder of read access views every account. That premise ended when the grants became the truth: identityapp.GrantPolicy limits a view plugin to its grants under the policy, the default, so a plugin installed later reads nothing there too. Widened on 2026-11-21 at the coordinator's request, after the alpha.2 release notes drafting found the gap.

## Source

> S4c: attention item `view-plugin-ungranted` (decisions -> missing settings, D-196): "<instance> can't read anything until you grant it accounts"; one action: grant the whole household, or open Access to pick. Raised for every installed view plugin with no D-209 grant while authz=spicedb; cleared when a grant exists; same sweep as model-route-broken; built-in UI exempt. CLI install summary for a view plugin: "grant it accounts in Settings -> Access".

Source: design-session answers for the permissions plan, 2026-11-13 (later)
