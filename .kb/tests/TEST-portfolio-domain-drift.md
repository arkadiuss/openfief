---
title: Drift is the weight held minus the target, exact and rounded half-even
status: active
sourceFile: internal/portfolio/drift_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:17.188Z'
id: TEST-portfolio-domain-drift
type: test
---
Table tests of the pure DriftOf: held above and below the targets, a target not held, rounding to two places, an instrument outside the model counted as other, and the refusals of a portfolio worth nothing or a negative holding.