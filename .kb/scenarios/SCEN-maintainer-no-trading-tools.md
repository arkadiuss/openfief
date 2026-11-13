---
title: The maintainer asks a model with no tools and applying trades nothing
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:24.824Z'
id: SCEN-maintainer-no-trading-tools
type: scenario
---
Given a maintainer run, when the request reaches the model provider, then it holds a system message and the figures and no tool at all; and when the owner applies a proposal, then only the model portfolio's weights change.