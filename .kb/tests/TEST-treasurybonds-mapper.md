---
title: Each bond lot becomes a deposit and a buy on its purchase date, keyed by series and day, with its terms once
status: active
sourceFile: internal/ingest/treasurybonds/mapper_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:41.012Z'
id: TEST-treasurybonds-mapper
type: test
verification_scope: unit
---
Golden-style comparison of Mapper.Map's transactions for three synthetic lots (two EDO0536, one TOS1029) in account acc-bonds: each is a deposit of quantity x 100 PLN against equity and a buy of the bonds for that cost, both on the purchase date, keyed lot:<series>:<date>:deposit and :buy under treasury-bonds:acc-bonds, every transaction valid; terms come once per series. Re-entering a lot with another quantity keeps its keys. Refusals: no account, a wrong series, a date after today, one series and day twice, one series with two rates; a refused batch maps nothing. TestInstrument checks the derived ISINs are valid and stable.