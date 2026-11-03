---
title: The portfolio schema stores kinds, sleeves and which account is in which sleeve
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T11:49:29.461Z'
id: SCEN-portfolio-pg-sleeves
type: scenario
---
Given a freshly migrated portfolio schema, then the six built-in kinds are stored; a direct change of a kind's default role is refused; a sleeve of an unknown kind, or with an override that doesn't match its role, is refused; and an account can be placed only in a stored sleeve.