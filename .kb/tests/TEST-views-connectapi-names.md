---
title: Every instrument the read API names carries its display name
status: active
sourceFile: internal/views/connectapi/names_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T06:59:00.489Z'
id: TEST-views-connectapi-names
type: test
---
REST over the real transcoder with a fake names port: holdings name the ETF and leave the bond unnamed, GetInstrument carries the name, a failing names port fails an answer with instruments and not one without.