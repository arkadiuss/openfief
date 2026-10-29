---
title: The UI presents its plugin token to the AuthService and the session to the read API
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:54.800Z'
id: SCEN-ui-core-credentials
type: scenario
---
Given the built-in UI running as a view plugin, when it calls the core's AuthService, ModelPortfolioService, BondLotService or StatementService, then each call carries the plugin's projected token read from its token file at that moment, also after it rotates, and nothing else: no build has a fixed token (D-84); when it calls the read API, each call carries the owner's session token as the bearer token and never the plugin's token.