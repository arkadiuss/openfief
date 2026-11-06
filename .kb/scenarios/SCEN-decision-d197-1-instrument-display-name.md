---
title: An instrument shows the first delivered name until renamed
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T06:59:02.863Z'
id: SCEN-decision-d197-1-instrument-display-name
type: scenario
---
Given IBKR delivers SYNTHETIC WORLD ETF for an ISIN and XTB later another name, when the read API and the UI show the instrument, then it is named SYNTHETIC WORLD ETF before its ticker and ISIN; after the owner renames it, the owner's name shows.