---
title: The UI draws the model portfolio while it is on and hands the owner's model to the core
status: active
sourceFile: internal/views/ui/model_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T00:07:41.781Z'
id: TEST-ui-model
type: test
verification_scope: integration
---
The UI over fake AuthService, ModelPortfolioService and read API servers: with the model on, the overview's chart carries the model's backtest as a second series aligned with the time-weighted return's days, and a table gives the model's 20% drawdown with its dates, a target that never fell and an index with no history; the settings page shows an empty model with blank rows, saves an instrument and an index through the core with the plugin's token and the owner's session, reads them back, and shows the core's refusal of weights not summing to 100 above the rows as posted; a read session gets 403 and a malformed form never reaches the core.