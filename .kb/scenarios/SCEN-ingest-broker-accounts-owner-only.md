---
title: Only the owner's account service lists broker accounts
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:14.978Z'
id: SCEN-ingest-broker-accounts-owner-only
type: scenario
---
Given broker accounts a connector reported, when the owner's session asks the account service, then it lists them, all or unmapped; a read session is refused, and the read API's ListBrokerAccounts answers UNIMPLEMENTED over Connect and REST, naming none.