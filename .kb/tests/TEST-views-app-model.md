---
title: The model portfolio query rebases the backtest on go-live and says nothing while it is off
status: active
sourceFile: internal/views/app/model_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:32:23.876Z'
id: TEST-views-app-model
type: test
verification_scope: unit
---
Over the views fakes and a fake report of 100, 105, 110, 99, 121 from 2026-02-27: with go-live on 2026-03-01 the line is 100, 90, 110 through the latest valued-through date, a later from restarts at 100, an earlier to ends there, and with nothing recorded the line is the whole backtest through 9999-12-31; an off model, a core built without the model, a failing report and a reversed range are refused with ErrModelOff or ErrInvalidRange.