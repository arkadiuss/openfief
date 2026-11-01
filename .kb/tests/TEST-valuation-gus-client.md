---
title: GUS CPI is read exactly, final months only, from the right indicator for the year
status: active
sourceFile: internal/valuation/gus/client_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T03:56:13.412Z'
id: TEST-valuation-gus-client
type: test
verification_scope: integration
---
hypert replays of GUS's public SDP API recorded on 2026-10-29 (no key, Set-Cookie dropped): 2025 from indicator 1832 gives twelve readings, September's 102.9 as 0.029, matching EDO1133's published 4.40% from November 2025 with its 1.50% margin; 2026 from indicator 2728 gives January to September; 2027 answers 404 and gives none. Stub tables check that preliminary, estimated, forecast or missing values end the year, a cumulative period is skipped, and another year, month-on-month presentation, duplicate months, text or zero values and HTTP errors fail.