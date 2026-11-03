---
title: The portfolio registry saves kinds and sleeves only when the stored kinds allow them
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T11:49:24.824Z'
id: SCEN-portfolio-registry-checks-sleeves
type: scenario
---
Given the built-in kinds and the sleeve core, when a sleeve of an offensive kind is saved defensive without the override, or of an unknown kind, or a kind's default role is changed, then each is refused and nothing is stored; a sleeve with no role takes its kind's default role.