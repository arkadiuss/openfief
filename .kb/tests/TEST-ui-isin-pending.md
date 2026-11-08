---
title: A bond series under its provisional ISIN carries one pending pill wherever its ISIN shows
status: active
sourceFile: internal/views/ui/isinpending_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-08T02:00:37.318Z'
id: TEST-ui-isin-pending
type: test
---
Serves the built-in UI over its fakes with the account at the treasury bond registry, which also holds 20 COI1130 bonds under the provisional ISIN PLCOI1130004 that the fake read API marks isin_provisional, and counts the pending pill with its tooltip on the holdings behind the PLN bar, the account's holdings and open lots tabs and the instrument page: one each, right after the ISIN on the instrument page; none on a bond under a registry ISIN, a listed instrument, or a broker account's lots. A false pass would be a pill shown by the ISIN's form rather than the read API's mark, which the fakes' registry-like EDO1137 ISIN guards against.