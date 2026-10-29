---
title: IBKR's mark prices become cross-check prices, never prices
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T06:24:27.154Z'
id: SCEN-ibkrflex-cross-check-prices
type: scenario
---
Given the synthetic Flex statement opening.xml, whose OpenPositions rows give TSTA a markPrice of 101.25 USD and TSTB 106.5 EUR on report date 2026-03-06, when it is mapped in either mode, then the mapping holds one cross-check price per instrument on that date and the delivery carries them in cross_check_prices with no prices; given LOT rows whose marks disagree and no SUMMARY row, then the instrument gets no cross-check price and a skipped problem; given an unreadable or zero markPrice, then a skipped problem, and the holding is still asserted.