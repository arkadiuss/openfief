---
title: Applying a reconcile plan as Ghostfolio does and planning again yields an empty plan
status: active
sourceFile: internal/views/ghostfolio/plan_property_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:14.421Z'
id: TEST-views-ghostfolio-plan-properties
type: test
verification_scope: unit
---
A property test over 500 cases from a seeded splitmix64 generator: random desired activities in one to three owned accounts, some missing, edited, copied, moved to another owned account or left as they should be, plus stray activities, missing owned accounts and the owner's own activities in a foreign account. A model of Ghostfolio applies the plan in Syncer's order, skipping imports its duplicate rule would skip; then a second plan must be empty, the owned accounts must hold exactly the desired activities, and the foreign activities must be unchanged. The model fails the test on any delete outside the owned accounts.