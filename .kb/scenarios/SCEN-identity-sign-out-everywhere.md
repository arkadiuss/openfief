---
title: The owner signs out every session at once
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:11.762Z'
id: SCEN-identity-sign-out-everywhere
type: scenario
---
Given owner sessions on two hosts and a demo session, when the owner signs out everywhere from an owner session, then every session not yet revoked is revoked, the caller's own included, an earlier revocation keeps its time, and logging in again works; a read session can't sign out everywhere. No passkey check is needed, since it only takes ways in away.