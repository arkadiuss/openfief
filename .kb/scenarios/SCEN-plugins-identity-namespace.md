---
title: Only plugin ServiceAccounts in openfief are plugin identities
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:13.956Z'
id: SCEN-plugins-identity-namespace
type: scenario
---
Given the authenticated usernames system:serviceaccount:openfief:plugin-sample, system:serviceaccount:default:plugin-sample and system:serviceaccount:openfief:operator, when each is mapped to a plugin, then only the first maps, to the plugin sample. Assumes the username comes from a TokenReview of a projected token with audience openfief-core.