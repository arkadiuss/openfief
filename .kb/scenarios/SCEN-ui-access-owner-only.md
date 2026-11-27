---
title: A read session never sees the Access page or its link and is refused every form of it
status: active
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
id: SCEN-ui-access-owner-only
type: scenario
---
Given a read-scoped session, the navigation has no Access link and GET /access and every Access form answer 403 without reaching the core; given the owner, under SpiceDB or under the policy, the page never says that grants have no effect.