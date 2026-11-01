---
title: The owner acknowledges and resolves only what the rules allow
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:09.484Z'
id: SCEN-attention-owner-actions
type: scenario
---
Given an open item, when the owner acknowledges it, it is acknowledged; when the owner resolves an import's item it is resolved, but resolving an unmapped broker account's item is refused because it resolves when the account is mapped, and a read session can do neither.