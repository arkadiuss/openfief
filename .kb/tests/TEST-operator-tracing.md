---
title: Each reconcile runs in an operator.reconcile span naming only the controller
status: active
sourceFile: internal/plugins/operator/tracing_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T03:05:55.534Z'
id: TEST-operator-tracing
type: test
---
traced wraps a reconcile.Func; a successful reconcile returns its result and runs with a span in its context, a failed one returns its error wrapped with the controller's name. Both produce operator.reconcile spans under the test's root whose only attribute is the controller, the failure an error without a description.