---
title: The core serves the read API over a real database with script tokens
status: active
sourceFile: cmd/core/core_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T14:32:46.383Z'
id: TEST-core-read-api
type: test
verification_scope: integration
---
Runs under task test:pg on a fresh, empty database. The core starts as serve does: it migrates every context, stores default sleeves, records the bond basis and starts its River workers. Part of the demo dataset is loaded the way a separate loader process would, through a ledger whose import hook enqueues revaluations, with synthetic FX rates. A token from the token issue command reads /healthz and /readyz (200), GET /v1/accounts and the XTB account's holdings, decoded strictly into the protobuf messages, and /v1/valuation-status until the core's own workers value the loaded accounts through today. No token, an unknown token and, after token revoke, the revoked token get 401 with a Bearer challenge and no data; token list never shows the secret. Shutdown then drains the HTTP server and stops the workers without error.