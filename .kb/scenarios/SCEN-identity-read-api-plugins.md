---
title: A view plugin reads the read API with its own projected token
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T01:36:15.060Z'
id: SCEN-identity-read-api-plugins
type: scenario
---
Given the Ghostfolio view running as plugin-ghostfolio with the view capability, when it calls the read API with its projected ServiceAccount token for audience openfief-core, then the core's verifier, finding no session or script token, asks the AuthService's caller check and accepts it; given a connector plugin's token or an unknown token, then the call is refused as unauthenticated; given a TokenReview that cannot be made, then the call fails as internal.