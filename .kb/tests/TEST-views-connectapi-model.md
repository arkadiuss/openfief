---
title: The read API serves the model portfolio as its contract says and 404 while it is off
status: active
sourceFile: internal/views/connectapi/model_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:32:25.823Z'
id: TEST-views-connectapi-model
type: test
verification_scope: integration
---
A generated Connect client and REST requests through the real transcoder over a fake report of a 60/40 ETF and FAKE20 model: the answer carries the targets, the backtest rebased to 100 on go-live, both backtest bounds, the model's and each target's drawdown and a target's missing FX rate; REST validates against the generated OpenAPI and a later from restarts at 100; an off model and a core without one answer not_found, no model portfolio, naming nothing; the protobuf enums match the domain's numbers.