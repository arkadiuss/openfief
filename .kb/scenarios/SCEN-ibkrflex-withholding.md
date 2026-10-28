---
title: Withholding tax rows are booked with their dividend
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:23.842Z'
id: SCEN-ibkrflex-withholding
type: scenario
---
Given a Dividends row of 7.5 USD with actionID 7000101 and a Withholding Tax row of -1.13 USD with the same actionID, and a dividend of 1.8 EUR without actionID with two withholding rows of -0.27 and -0.03 EUR on the same ISIN and date, when the statement is mapped, then the first dividend is booked with 1.13 USD withheld and 6.37 USD cash, the second with 0.30 EUR withheld; a withholding row with no dividend, several dividends or a positive amount becomes an unsupported problem.