---
title: A host that returns reasoning with an empty reply shows its length in report.json and none of its text
status: active
origin:
  kind: agent
  ref: Design session approval of the empty-reply fix after bench traces (D-205, D-217), 2026-11-15
  recorded_at: '2026-11-15T12:26:15.363Z'
id: SCEN-askbench-traces-reasoning-length
type: scenario
---
Given a host that returns reasoning with a reply, when the bench traces the round, then report.json holds the reasoning's length in characters and none of its text.

## Context

Reasoning is never an answer and is not kept.