---
title: Creating a demo over the budget is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:14.718Z'
id: SCEN-tooling-demo-over-budget
type: scenario
---
Given a project where another demo already costs 8.3181 EUR a month gross and DEMO_BUDGET_EUR is 10, when task demo:up runs for a new demo whose server and IPv4 cost 8.3181 EUR, then it prints every line and the total of 16.6362 EUR or more, refuses, and creates nothing. Assumes the price list and inventory come from the Hetzner API; prices are rounded up.