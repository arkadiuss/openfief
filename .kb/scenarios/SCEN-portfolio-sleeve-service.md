---
title: The owner adds and changes kinds and sleeves through the core
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:34.272Z'
id: SCEN-portfolio-sleeve-service
type: scenario
---
Given an owner session in the built-in UI, when the owner adds the kind Gold as defensive, renames Satellite, adds the sleeve Gold bars of kind gold and makes Satellite defensive with the override, then the core stores gold, gold-bars and the changes; when the owner adds a sleeve whose role differs from its kind's without the override, or a name whose ID is taken, the core refuses it with the reason and stores nothing; a read session or another plugin is refused.