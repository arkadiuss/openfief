---
title: The tracing component turns tracing on for the core and the operator, and the bases keep it off
status: active
sourceFile: scripts/tests/tracing.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-02T03:05:57.619Z'
id: TEST-deploy-tracing-component
type: test
---
Builds docs/guides/tracing.md's own overlay with kustomize and checks with gojq that the ConfigMap holds the owner's endpoint and sampler, that the core and the operator read it through envFrom and take the OTLP headers from the optional Secret, that the core keeps its own settings, and that a NetworkPolicy lets the core reach the collector's pods on 4317 and 4318 alone; then that the core and operator bases set no tracing variable and the core's policy reaches no collector.