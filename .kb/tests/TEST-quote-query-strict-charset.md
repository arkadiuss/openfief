---
title: A quote query of dots, with two in a row, a leading or trailing dot or dash, a slash or an underscore is refused in the core, the tool and the connector
status: active
sourceFile: internal/ingest/quote_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #5 (2026-11-21)'
  recorded_at: '2026-11-21T18:46:39.457Z'
id: TEST-quote-query-strict-charset
type: test
---
Table tests of ingest.ParseQuoteQuery (BRK-B, SYNQ.XETRA and an ISIN taken; ., .., SYNQ..X, .SYNQ, SYNQ., -SYNQ, SYNQ/X and SYNQ_X refused), of the market price tool's arguments (.. and .SYNQ refused), of the read API's REST answer (400 for ..%2E and SYNQ..X, quoting nothing) and of the EODHD connector's LookupQuote (., .., SYNQ..X and SYNQ/X are InvalidArgument before any call to EODHD).