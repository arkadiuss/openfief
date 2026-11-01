---
title: A current xStation export in opening mode seeds opening lots and the opening cash
status: active
sourceFile: internal/ingest/xtb/opening_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:14.454Z'
id: TEST-ingest-xtb-opening
type: test
---
Builds synthetic current xStation exports with xtbtest (account 0000000, report generated 2026-06-30, two TESTA.PL positions, one with a float artifact in its open price, a TESTB.UK ETF quoted in another currency and a CFD) and maps them with xtb.Mapper in opening mode. Golden files fix the mapping with 2500.50 PLN of opening cash, with zero and with none entered. Other cases check that the same export, renamed, maps the same again, that nothing of Cash Operations or Closed Positions is read, and how open positions are keyed. A false pass would need the golden file itself to be wrong.