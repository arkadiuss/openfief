---
title: The owner's Grant the whole household on an ungranted view plugin's item makes the household grant
status: active
origin:
  kind: agent
  ref: 'design-session answers for the permissions plan, 2026-11-13 (later): attention item view-plugin-ungranted (D-196, D-209)'
  recorded_at: '2026-11-13T17:06:27.724Z'
id: SCEN-ui-inbox-grant-ungranted-view-plugin
type: scenario
---
Given an open view-plugin-ungranted item, the Inbox shows the owner a Grant the whole household button, which posts the instance and reaches the core as CreateGrant of the household to that view plugin, then returns to the Inbox saying the plugin can read; a refusal such as a plugin uninstalled meanwhile shows the core's reason, an empty instance is refused before the core is called, and a read session sees the item without the button and is refused the post. The grant is one audit entry by the owner, and the item clears at the next sweep.