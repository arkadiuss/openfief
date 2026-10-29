---
title: An in-cluster service lies outside openfief and is named only when the manifest asks
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:15.685Z'
id: SCEN-plugins-cluster-service-outside-openfief
type: scenario
---
Given a manifest asking for an in-cluster service, when the owner picks ghostfolio/ghostfolio:3333, then the Plugin requests exactly it and runs with it only once the grant approves exactly it; a pick in the openfief namespace, a malformed one, or one for a manifest that asks for none is refused.