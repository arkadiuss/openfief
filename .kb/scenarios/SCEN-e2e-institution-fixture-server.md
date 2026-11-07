---
title: The connectors read synthetic institution answers from the fixture server over HTTPS
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T22:44:06.877Z'
id: SCEN-e2e-institution-fixture-server
type: scenario
---
Given the fixture server started with a certificate and key and synthetic Flex and EODHD keys, when the IBKR Flex, EODHD and NBP clients the connectors use call it at its path for each service, then they parse a synthetic statement, synthetic listings and closes and synthetic table A rates, and a wrong key is refused the way the real service refuses it. Assumes no real account data is in its answers.