---
title: A sleeve with no accounts and a kind no sleeve has are deleted
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:13.194Z'
id: SCEN-portfolio-delete-unused-sleeves-and-kinds
type: scenario
---
Given a sleeve with accounts, a sleeve without and the kinds they have, when the owner deletes them, then the sleeve with accounts and a kind in use are refused, and the empty sleeve and then its now unused kind are deleted, built-in kinds too.