---
title: Provisional ISINs are built and told from registry ISINs; published data looks up a series' registry ISIN
status: active
sourceFile: internal/jurisdiction/pl/isin_test.go
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:11.774Z'
id: TEST-pl-isin
type: test
---
Table tests over synthetic series codes and ISINs: ProvisionalPrefix builds PL, the code and 00; ProvisionalSeries and IsProvisional accept only an ISIN whose national number is its series' code and filler, and refuse a synthetic registry ISIN, another country's code, a missing filler, an unknown bond type and a short string; Published.ISINOf returns a listed series' registry ISIN, one listed before its rates too, and ErrNotPublished otherwise. A false pass would be a registry-like ISIN read as provisional, which the core would then re-key.