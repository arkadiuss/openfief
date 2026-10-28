---
title: The ledger schema stores sleeve kinds and gives existing sleeves one
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:53:03.178Z'
id: SCEN-ledger-pg-sleeve-kinds
type: scenario
---
Given a database with sleeves stored before kinds existed, when the sleeve kinds migration runs, then the built-in kinds are stored and each defensive sleeve gets kind core and each offensive one kind satellite. Afterwards a sleeve naming an unstored kind is refused, an offensive override on a defensive sleeve is refused, and any update that changes a stored kind's default role is refused by the database itself.