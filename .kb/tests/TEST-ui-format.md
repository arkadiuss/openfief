---
title: The UI formats exact decimal amounts, shares, labels, dates and chart data
status: active
sourceFile: internal/views/ui/format_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:34.112Z'
id: TEST-ui-format
type: test
verification_scope: unit
---
Table tests of the UI's pure helpers: thousands grouping keeps sign and every decimal place, shares are percentages rounded half-even to one place with the kernel's Decimal.Quo and empty for a zero total, enum and valuation status labels, the page date from a query, the latest valued-through date or today, the chart data as exact strings, the browser origin from X-Forwarded-Proto and Host, the cookie attributes and the token lifetime bounds.