---
title: An xStation export maps to transactions and ticker entries, never keeping a transfer's comment
status: active
sourceFile: internal/ingest/xtb/mapper_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:18.487Z'
id: TEST-ingest-xtb-mapper
type: test
---
Builds synthetic xStation exports in the test with xtbtest (account 0000000, tickers TESTA.PL and TESTB.UK, operation IDs 9000000001 and up, a made-up sending bank) and maps them with xtb.Mapper. Golden files fix the whole mapping of an export with every R-24 row type, CFD rows and float artifacts, and the same export with inline strings and no Total rows. Other cases check that no part of a deposit's or withdrawal's comment reaches the mapping and the account number names only the broker account; that an opening batch, an unknown mode, a non-workbook, a file too large, a missing sheet, another account's file name and two currencies are refused; that a renamed file takes its currency from the summary; and that artifacts round away while a real fraction of a grosz is an ingest problem. A false pass would need the golden file itself to be wrong.