---
title: A request goes only to models the owner allowed for every tier it holds
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:19.606Z'
id: SCEN-models-data-tiers-deny-by-default
type: scenario
---
Given requests of messages tagged T0, TF, T1 and T2, the router sends one only to a model of its route that the owner listed for every tier in it; a tier with no models listed, a request with an untagged message and a request none of the route's models may receive are refused and recorded; a fallback never widens the tiers.