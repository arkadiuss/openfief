---
title: Running up twice creates nothing the second time
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:19.781Z'
id: SCEN-tooling-demo-idempotent-up
type: scenario
---
Given a demo whose server and firewall carry the labels app=openfief and env=demo, when task demo:up ENV=demo runs again, then no server or firewall is created, the firewall rules are replaced with the current DEMO_SSH_CIDRS, and an already attached firewall is not attached again.