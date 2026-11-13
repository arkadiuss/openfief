---
title: An installed view plugin without a grant raises an item under SpiceDB and clears when granted or uninstalled
status: active
origin:
  kind: agent
  ref: 'design-session answers for the permissions plan, 2026-11-13 (later): attention item view-plugin-ungranted (D-196, D-209)'
  recorded_at: '2026-11-13T17:06:26.925Z'
id: SCEN-identity-view-plugin-ungranted-attention
type: scenario
---
Given the core runs with SpiceDB authorization and a view plugin instance that is Ready, holds the view capability and serves no browser path has no grant in force, the periodic check raises one item of the kind view-plugin-ungranted, in the decisions group, naming the instance with the text "<instance> can't read anything until you grant it accounts"; a grant of the household, a sleeve or an account to it clears the item at the next sweep, and so does uninstalling the plugin; a plugin that is not Ready, is being removed, lacks the view capability or serves the browser path, as the built-in UI does, raises none, and the core running with the policy raises none; a plugin cache that has not synced makes the check fail instead of clearing items.