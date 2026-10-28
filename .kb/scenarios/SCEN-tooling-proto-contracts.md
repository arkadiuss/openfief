---
title: A breaking contract change fails the check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:37:17.928Z'
id: SCEN-tooling-proto-contracts
type: scenario
---
Given the contracts at the base revision, when a change keeps them compatible (such as adding a field) the breaking check passes; when it breaks them (such as changing a field's type) the check fails and names what broke; when the base has no contracts yet there is nothing to break and the check passes.