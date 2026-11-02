---
title: Every plugin container is sized and capped
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T20:55:32.303Z'
id: SCEN-operator-plugin-resources
type: scenario
---
Given any granted Plugin, when the operator builds its Deployment, then the plugin container requests 50m CPU and 64Mi of memory and is limited to 500m CPU and 256Mi of memory, whatever its manifest asks for, so a plugin that leaks memory is killed alone at its limit and one that spins is throttled instead of starving the node and the database next to it.