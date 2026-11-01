---
title: Events are published in order and resumed after a failure
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:15.848Z'
id: SCEN-attention-relay
type: scenario
---
Given three events, when the publisher fails on the second, the relay stops after the first, and the next run publishes the second and third in order; the log publisher writes the IDs, kind, state and reason but not the subject or detail; the PostgreSQL store lists unpublished events in order and records each publication once.