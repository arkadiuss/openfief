---
title: The ledger saves kinds and sleeves only when the stored kinds allow them
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:52:53.186Z'
id: SCEN-ledger-app-kinds
type: scenario
---
Given the built-in kinds are stored, when the owner saves a custom kind or renames a built-in one, then it is stored; when the owner changes a stored kind's default role, then it is refused and nothing changes. When a sleeve is saved with no role, then it is stored with its kind's default role; when it names a kind that is not stored, or is offensive in a defensive kind without the explicit override, then saving fails and the stored sleeves are unchanged.