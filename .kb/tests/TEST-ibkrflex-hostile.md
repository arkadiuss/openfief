---
title: Hostile and malformed Flex input is refused within limits
status: active
sourceFile: internal/ingest/ibkrflex/hostile_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:16.720Z'
id: TEST-ibkrflex-hostile
type: test
verification_scope: unit
---
Malicious statement files from DESIGN section 3.5 are fed to Map: a billion-laughs DTD, an external entity, an undeclared entity, a non-UTF-8 encoding, truncated and empty input, a FlexStatementResponse error document, an unreadable toDate, and input over each of the size, depth, element count, attribute count and attribute length limits. Each must fail with ErrMalformed or ErrTooLarge and return no mapping. Input exactly at every limit is accepted, an unknown ingest mode is refused, and a failing reader surfaces as ErrMalformed wrapping its error.