---
title: The same seed and date give the same demo dataset
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:35.374Z'
id: SCEN-demo-deterministic
type: scenario
---
Given seed 20261028 and the date 2026-10-28, when the dataset is generated twice, then both are equal down to every statement byte, transaction and price; when the seed differs, then the prices differ; when the date differs, then the history differs.