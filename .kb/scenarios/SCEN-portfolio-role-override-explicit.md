---
title: A role other than the sleeve kind's needs the owner's explicit override, either way
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:45.208Z'
id: SCEN-portfolio-role-override-explicit
type: scenario
---
Given the built-in kinds, when a sleeve of kind satellite is saved defensive without the role override, or a sleeve of kind bonds offensive without it, then it is refused with ErrRoleOverride and nothing is stored; with the override both are stored; a sleeve whose role is its kind's is refused with the override set; and PostgreSQL refuses the same rows, and gives sleeves stored before D-71 the override where their role differs.