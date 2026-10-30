---
title: The IBKR Flex plugin reports its Flex accounts instead of mapping them itself
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T11:36:04.324Z'
id: SCEN-ingest-ibkrflex-by-broker-account
type: scenario
---
Given the ibkr-flex plugin with no config and a synthetic Flex statement of U0000000 and U0000001, when the core syncs it or uploads the statement, then every transaction names its IBKR account and no OpenFief account, keyed by the source ibkr-flex, and both accounts are listed in broker_accounts.