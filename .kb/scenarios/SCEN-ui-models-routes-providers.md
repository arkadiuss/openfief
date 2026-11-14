---
title: The owner sees the providers' terms and edits a role's route with broken entries marked
status: active
origin:
  kind: agent
  ref: the MS1 brief, from D-205 and D-196, 2026-11-13
  recorded_at: '2026-11-14T00:24:12.832Z'
id: SCEN-ui-models-routes-providers
type: scenario
---
Given providers and routes, when the owner opens the Models page then each provider shows where its models run, what it keeps, its prices and limits; an entry naming a missing provider or a non-provider plugin is marked broken with a fix, and the owner adds, removes and moves entries, each change reaching the core whole, with the core's refusal shown.

## Context

D-205 has providers declare their location and retention and routes name provider/model references. The page shows what a provider declares as information, edits routes one change at a time through the core, and marks a broken entry as the model-route-broken item does, linking the Inbox item to the route.

## Source

> A plugin ... declares where its models run and the provider's retention terms, lists models ... The core ... gives each role ... a primary model and ordered fallbacks as provider/model references.

Source: the MS1 brief, from D-205 and D-196, 2026-11-13
