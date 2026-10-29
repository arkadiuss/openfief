---
title: The read API lists the installed plugin instances from the operator's resources
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:30.121Z'
id: SCEN-views-installed-plugins
type: scenario
---
Given Plugin resources in the openfief namespace, when a client calls ListPlugins or GET /v1/plugins with a valid token, then every installed instance comes back in name order with its image, digest, effective and requested capabilities, trust level, signer and phase; when the core cannot read the cluster the call is UNAVAILABLE; nothing in the API changes a plugin or grant.