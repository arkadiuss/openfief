---
title: PostgreSQL keeps the charts' start per portfolio and account, clears it, and forgets a deleted account's
status: active
sourceFile: internal/valuation/postgres/chartstart_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-02T08:22:18.116Z'
id: TEST-valuation-pg-chart-start
type: test
---
Runs valuation's store against a disposable PostgreSQL with every migration applied (task test:pg). Nothing stored reads as the zero date, the portfolio's start and two accounts' starts are kept apart and overwritten in place, saving the zero date clears only the portfolio's, and ForgetAccount removes a deleted account's start while another account's stays.