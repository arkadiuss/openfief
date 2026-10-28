---
title: The stored model portfolio is off until saved and is replaced whole
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:22.022Z'
id: SCEN-portfolio-model-store
type: scenario
---
Given a fresh database, when the model portfolio is read, then it is off with no targets; when a model is saved, then whether it is on and all its targets are replaced in one transaction and read back in the owner's order, and a target breaking a table check refuses the whole save.