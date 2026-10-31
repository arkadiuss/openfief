---
title: The eodhd plugin proposes the one ISIN listed for a ticker's code on its market's exchange
status: active
sourceFile: plugins/eodhd/lookup_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:32.881Z'
id: TEST-plugins-eodhd-lookup
type: test
---
Calls the eodhd connector's LookupInstruments handler over a fake EODHD source: a code listed once on the supported exchange of its country is proposed with its name, an ambiguous code, an unknown one and a country with no supported exchange are not, and only supported exchanges are searched with the API key. A missing key, EODHD failing and more than 100 tickers are refused with their codes.