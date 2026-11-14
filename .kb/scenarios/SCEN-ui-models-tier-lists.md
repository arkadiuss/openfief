---
title: The owner allows models per data tier in a grid and denied is the default
status: active
origin:
  kind: agent
  ref: the MS1 brief, from D-206, 2026-11-13
  recorded_at: '2026-11-14T00:24:10.622Z'
id: SCEN-ui-models-tier-lists
type: scenario
---
Given providers with models, when the owner opens the Models page then each tier is explained and each model has an unticked box for each tier until ticked, saving the grid sends one change for each tier whose models changed, unticking everything clears the tiers, and a route whose models may not receive what the role sends is flagged.

## Context

D-206 makes the per-tier lists the owner's, with deny by default. The page shows the four tiers in plain words and a grid of models and tiers; a model a tier allows that no provider offers any more stays listed so it can be unticked.

## Source

> The owner lists, per tier, the models allowed to receive it, and the router sends a request only to a model allowed for every tier in it, denying by default.

Source: the MS1 brief, from D-206, 2026-11-13
