---
title: An install gets the built-in kinds and three default sleeves once
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:34:33.587Z'
id: SCEN-portfolio-default-kinds
type: scenario
---
Given a fresh install with no kinds or sleeves, when the defaults are computed, then all six built-in kinds and the sleeves Bonds, Core and Satellite are missing; given an install with everything stored, then nothing is missing; given an install where the owner renamed the core kind and moved the core sleeve to another kind, then only the absent defaults are missing and the owner's changes are kept.