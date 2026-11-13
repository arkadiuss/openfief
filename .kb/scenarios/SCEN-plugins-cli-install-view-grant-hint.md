---
title: The install of a view plugin tells the owner to grant it accounts in Settings → Access
status: active
origin:
  kind: agent
  ref: 'design-session answers for the permissions plan, 2026-11-13 (later): attention item view-plugin-ungranted (D-196, D-209)'
  recorded_at: '2026-11-13T17:06:28.476Z'
id: SCEN-plugins-cli-install-view-grant-hint
type: scenario
---
Given openfief plugin install of a plugin with the view capability, the output ends with "If OpenFief runs with SpiceDB authorization, grant it accounts in Settings → Access."; given a plugin with only the connector capability, the output has no such line.