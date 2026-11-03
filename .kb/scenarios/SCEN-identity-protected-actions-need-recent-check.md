---
title: A protected action needs a passkey check in the last five minutes
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:02.200Z'
id: SCEN-identity-protected-actions-need-recent-check
type: scenario
---
Given an owner session that a passkey signed in, or that confirmed with a passkey, when five minutes or more have passed since that check and the owner asks to create a script token, issue an enrollment link, begin creating a passkey with the session or remove a passkey, then the core refuses with a passkey-check-required error and changes nothing; within five minutes the same requests succeed. A demo session can't take any of them in any case.