---
title: The IBKR Flex mapper names entries by IBKR account and resolves to what a configured mapper books
status: active
sourceFile: internal/ingest/ibkrflex/brokers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:33.116Z'
id: TEST-ingest-ibkrflex-brokers
type: test
verification_scope: unit
---
Maps both synthetic Flex fixtures in both modes with ByBrokerAccount: every entry is grouped by its synthetic IBKR account in no account with the ibkr-flex source, every account is reported, and resolving U0000000 to acc-ibkr gives exactly the transactions and assertions, keys included, of the mapper configured with that account, while U0000001 stays unmapped. A false pass would need both paths to drop the same rows, which the golden tests rule out.