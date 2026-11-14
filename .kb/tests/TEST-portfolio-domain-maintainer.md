---
title: The maintainer's prompt, schema, answer checks and proposal rules
status: active
sourceFile: internal/portfolio/maintainer_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:17.887Z'
id: TEST-portfolio-domain-maintainer
type: test
---
Table tests of BuildPrompt (figures as weights, instructions with none), the strict ParseAnswer and CheckChanges rules, the figures filled behind each change, and Proposal.Apply with stale and off models.