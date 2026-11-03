---
title: The refusal tells a view plugin to ask for a passkey check
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:07.162Z'
id: SCEN-identity-passkey-check-required-detail
type: scenario
---
Given a view plugin calling the AuthService with an owner session whose last passkey check is stale, when it asks for an action D-183 protects, then the call fails with FAILED_PRECONDITION and a PasskeyCheckRequired error detail naming the five-minute window, which no other refusal carries.