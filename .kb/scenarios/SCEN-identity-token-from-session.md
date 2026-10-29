---
title: Script tokens are created from an owner session
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:57.629Z'
id: SCEN-identity-token-from-session
type: scenario
---
Given an owner session, when the owner creates a script token, then they get a read-only token whose secret is shown once; without a session, or with a read session, creation is refused.