---
title: The owner's tickers page lists proposals and hands confirmations to the core
status: active
sourceFile: internal/views/ui/tickers_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:42.812Z'
id: TEST-ui-tickers
type: test
---
Serves the built-in UI against fake core services over HTTP: the tickers page shows each pending ticker with its proposal filled in or an empty field, escaped, for the owner only and linked from the menu; a confirmation reaches the core trimmed and in upper case and the page says how many rows it booked; a forged count is not shown; the core's refusal and an overlong ISIN are shown on the page; and an upload whose rows wait links the page.