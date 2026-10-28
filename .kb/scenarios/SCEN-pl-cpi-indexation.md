---
title: After the first period, rates follow CPI or the NBP reference rate plus the margin
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:25.050Z'
id: SCEN-pl-cpi-indexation
type: scenario
---
Given an EDO series with a 2% margin and a CPI reading of 4.9% for its second year, when the second year's rate is computed, then it is 6.9%; given a CPI of -1%, then the rate is the 2% margin; given ROR with the NBP reference rate at 5.75% and no margin, then its next monthly rate is 5.75%; given a negative reference rate, then it counts as zero; given TOS, then every year keeps the first-year rate. The CPI reading is the one for the month two months before the period starts.