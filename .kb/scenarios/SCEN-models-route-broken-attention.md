---
title: A broken route entry raises an item naming the route and the instance and clears when fixed
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T14:45:42.458Z'
id: SCEN-models-route-broken-attention
type: scenario
---
Given a route whose primary names a provider instance no plugin has, or one that is a connector, a call falls through to the next model and one attention item of the kind model-route-broken, in the decisions group, names the role and the instance; the periodic check raises the same item for every broken entry of a route in use, clears it when the entry is fixed or removed, and raises none when the switch or the role's toggle is off; given a provider that is only down, the call falls through and no item is raised.