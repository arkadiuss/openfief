---
title: Destroying a demo needs its name
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:24.235Z'
id: SCEN-tooling-demo-confirmed-down
type: scenario
---
Given a demo named demo, when task demo:down runs without CONFIRM or with CONFIRM set to another name, then nothing is deleted; when CONFIRM=demo, then only resources labelled app=openfief,env=demo are deleted, the server before its firewall.