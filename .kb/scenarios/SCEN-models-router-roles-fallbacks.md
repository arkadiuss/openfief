---
title: A role's models are tried in order and an unavailable one passes on
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:16.756Z'
id: SCEN-models-router-roles-fallbacks
type: scenario
---
Given a route of a primary and fallbacks, when the primary is unavailable the next model answers with its own entry; when every model is unavailable the call fails as unanswered; when a provider rejects the request the call ends without a fallback; routes round-trip through the owner's settings.