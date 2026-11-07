---
title: A changed row for the same IBKR transaction ID is a new version and the old one stays superseded
status: active
tags:
  - acceptance:p10-3
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:36.540Z'
id: SCEN-e2e-correction-new-version
type: scenario
---
Given the IBKR account holding the week statement deposit of 1000 GBP, when the fixture serves the corrected week, whose deposit under the same transaction ID is 1100 GBP, and the instance syncs, then the deposit has a second version, the first kept and marked superseded, no other transaction changes, and the read API shows the deposit at 1100 GBP. Runs on both kind and k3d.