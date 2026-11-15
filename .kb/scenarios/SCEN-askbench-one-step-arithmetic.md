---
title: A number one step from two shown numbers passes the numbers check, and a longer chain does not
status: active
origin:
  kind: agent
  ref: Design session decisions on A19 as a false-premise case and one-step arithmetic in the numbers check, 2026-11-14
  recorded_at: '2026-11-14T22:23:40.169Z'
id: SCEN-askbench-one-step-arithmetic
type: scenario
---
Given an answer that writes the difference, the sum or the percentage change of two numbers of the shown cards, when it is graded, then it passes when it equals the exact result rounded half-up to the places it writes (1,966.85 and 1,967 for a change of 1,966.85; 3.47% and 3.5% for a change of 3.4678%), and fails at 1,966.00, 1,966, 2,000, 3.46%, a result of three numbers, a sum of a sum, and a number from outside the cards.