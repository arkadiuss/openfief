---
title: The core takes an owner's uploaded Flex statement, hands it to the bound connector and records the batch
status: active
sourceFile: cmd/core/uploads_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T04:59:56.577Z'
id: TEST-core-statements
type: test
---
PostgreSQL test through the real core (task test:pg): after a passkey login, the synthetic IBKR Flex statement activity.xml is streamed over HTTP/1.1 to StatementService, refused UNAVAILABLE before the core found a connector, then handed to an in-process connector that maps it with ibkrflex like the ibkr-flex plugin: 10 new transactions, 4 assertions, 12 problems; uploading it again leaves all 10 unchanged. Both batches are recorded as completed uploads of flex-main, the deposit is in the account under a key scoped to the instance, and an unknown session, an unbound account and a CSV are refused.