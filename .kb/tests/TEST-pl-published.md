---
title: Published rates are looked up by series and their attribution must name everything the licence asks for
status: active
sourceFile: internal/jurisdiction/pl/published_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T08:05:53.776Z'
id: TEST-pl-published
type: test
---
TestPublishedRatesOf looks synthetic series up in pl.Published, an indexed one with its margin, a fixed-rate one without, and one the data doesn't list, which fails with ErrNotPublished. TestAttributionValidate removes each part of a synthetic attribution in turn and expects ErrInvalidAttribution. A false pass would be a lookup that matched the wrong series or a validation that ignored a missing part.