---
title: Only the owner reads or replaces the model portfolio, through the view plugin
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:20.217Z'
id: SCEN-portfolio-model-entry
type: scenario
---
Given the built-in UI calling the core with its plugin token and the owner's session, when it reads or saves the model portfolio, then an owner-scoped session on an allowed origin may, a read session or another origin gets permission denied, an invalid session unauthenticated, a model that breaks a rule or names a malformed ISIN invalid argument, and nothing is stored on any refusal.