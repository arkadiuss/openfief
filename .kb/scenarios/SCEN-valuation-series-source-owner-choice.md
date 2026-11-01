---
title: The owner switches one instrument's source
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:22.937Z'
id: SCEN-valuation-series-source-owner-choice
type: scenario
---
Given an instrument priced by two sources and held by one of two accounts, when the owner chooses the other source, then only that account is revalued from its first transaction and the prices come from the chosen source alone; choosing it again revalues nothing, and a source without any price of the instrument is refused.