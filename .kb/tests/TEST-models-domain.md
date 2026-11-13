---
title: The models domain decides the switch, routes, tiers and plan
status: active
sourceFile: internal/modelrouter/modelrouter_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T11:27:58.121Z'
id: TEST-models-domain
type: test
---
Table tests of the pure domain: the switch and its toggles, references and routes, the tier set and policy, MakePlan's refusals and candidates, request validation and the entries' capabilities, and that no error quotes a request.