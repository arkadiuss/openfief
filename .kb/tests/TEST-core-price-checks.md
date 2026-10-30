---
title: The core serves PriceCheckService for the owner and lists checks against the latest price
status: active
sourceFile: cmd/core/pricechecks_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T02:58:22.288Z'
id: TEST-core-price-checks
type: test
---
Through the real core on PostgreSQL with a passkey owner session, stores Friday's close and a Sunday IBKR mark 3% above it, checks through the read API that the mark is flagged at 1% against Friday's price, acknowledges it through PriceCheckService and sees it acknowledged, sets the instrument's threshold to 5% and sees it matched, then checks that malformed thresholds, a caller that is not a view plugin, an unknown session, acknowledging a matched check and acknowledging no check are each refused with their own code.