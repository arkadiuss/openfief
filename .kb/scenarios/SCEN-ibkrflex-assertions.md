---
title: Flex open positions and cash report become balance assertions on the statement's last day
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:26.219Z'
id: SCEN-ibkrflex-assertions
type: scenario
---
Given a synthetic Flex statement ending 2026-03-13 with a SUMMARY open position of 20 units of one ETF, two LOT rows of 3 and 1.5 units of a stock and no summary row for it, and a cash report with ending cash in USD, EUR and the base-currency summary, when it is mapped in either mode, then it yields holding assertions of 20 and 4.5 units and cash assertions per real currency dated 2026-03-13, ignores the base-currency summary, and turns a short position and a futures position into unsupported problems.