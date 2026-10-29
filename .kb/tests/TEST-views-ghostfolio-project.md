---
title: Each transaction kind becomes the right Ghostfolio activities, with derived UUIDs and bond prices
status: active
sourceFile: internal/views/ghostfolio/project_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:10.591Z'
id: TEST-views-ghostfolio-project
type: test
---
Pure table tests of Project on one synthetic account (IETEST000026 listed as YAHOO ETF1.L, PLTEST000010 EDO0136 as a treasury bond): a buy with its fee, a fee in another currency split into a FEE with a #fee:EUR comment, a sell at the proceeds' unit price, an opening lot dated when the lot was opened, a dividend of quantity 1 with the withheld tax as fee, a fee against the PLN fee asset, cash movements with no activity, an unlisted instrument as a problem, 100/3 rounded to 33.3333333333333, and a bond as a MANUAL asset priced 100, 100.027 and 100.03 from valuations divided by the quantity held. TestDerivedUUIDs pins the UUID version 5 values against Python's uuid.uuid5; TestProjectAccounts the per-account owned accounts and order; TestProjectRejects a 19-digit cost and a 2000-character key. Comments are checked exactly, so a missing transaction key fails.