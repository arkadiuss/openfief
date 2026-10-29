---
title: The UI shows the model portfolio while it is on and lets the owner enter it
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T00:07:46.198Z'
id: SCEN-ui-model-portfolio
type: scenario
---
Given the model portfolio on, when the overview is shown, then the model's backtest is the chart's second line on the same days, 100 on go-live, with a table of the model's and each target's max drawdown and dates; given an owner session on the settings page, when the owner saves the form, then the targets go to the core's model portfolio service with the plugin's token, a refusal shows its reason with the rows as posted, and a read session is refused.