---
title: A week-old series without ISIN or terms raises a decision item
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:34.132Z'
id: SCEN-attention-bond-series-incomplete
type: scenario
---
Given an account has held a treasury bond series for seven days or more and it is still under its provisional ISIN, without stored terms, or both, when the attention checks run, then one item is open per ISIN, among the decisions, saying what is missing since the first day held; a series held less stays silent, and the item resolves once both arrive or the series is no longer held.