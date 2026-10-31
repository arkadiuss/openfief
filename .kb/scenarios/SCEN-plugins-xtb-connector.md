---
title: The xtb plugin serves uploads only
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:55:05.522Z'
id: SCEN-plugins-xtb-connector
type: scenario
---
Given the xtb plugin served with the SDK, when the core streams an export to ImportFile, then it answers with the delivery; when the core syncs it, it delivers nothing and keeps the watermark; and its manifest asks for statement files with no egress and no Secret.