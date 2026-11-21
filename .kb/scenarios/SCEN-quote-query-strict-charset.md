---
title: A quote query that could be a path segment never reaches the provider
status: active
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #5 (2026-11-21)'
  recorded_at: '2026-11-21T18:46:40.483Z'
id: SCEN-quote-query-strict-charset
type: scenario
---
Given an install with the EODHD connector, when a caller asks GET /v1/market-quotes/ with a query of dots alone, two dots inside, a leading or trailing dot or dash, a slash or an underscore, then the read API answers 400 with its fixed message and asks no connector, the Ask box's market price tool refuses the same arguments, and the connector refuses them as InvalidArgument without calling EODHD; BRK-B, SYNQ.XETRA and an ISIN are still looked up.