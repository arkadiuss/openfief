---
title: The real core stores the owner's charts' start and an account opened on its seeding day
status: active
sourceFile: cmd/core/chartstart_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-02T08:22:20.393Z'
id: TEST-core-chart-start
type: test
---
Starts the core on a disposable PostgreSQL with a passkey owner session (task test:pg). The AccountService creates an account marked as opened on its seeding day and the read API serves the mark; SettingsService SetChartStart stores the portfolio's and the account's start, which GetPerformance names as chosen_from over REST, an empty start clears the portfolio's only, and a malformed start, an unknown account and an unknown session are refused with their Connect codes.