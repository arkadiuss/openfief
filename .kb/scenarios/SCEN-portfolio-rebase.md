---
title: A series rebased on a date is 100 on its first day on or after that date
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:31.225Z'
id: SCEN-portfolio-rebase
type: scenario
---
Given the backtest's index and go-live, when the model's line is drawn from go-live, then the series from the first day on or after go-live is divided by that day's value and multiplied by 100, rounded half-even to 12 places, so both chart lines start at 100.