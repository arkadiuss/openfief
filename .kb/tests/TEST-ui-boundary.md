---
title: The UI imports only the kernel and the SDK, and its vendored libraries are the pinned releases
status: active
sourceFile: internal/views/ui/boundary_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:36.152Z'
id: TEST-ui-boundary
type: test
---
Parses every non-test Go file of internal/views/ui and fails on any OpenFief import other than the kernel and the SDK's generated clients and helpers, so the UI can only read the record through the public read API. Checks that New refuses a config without both clients, and that the vendored htmx and Chart.js files hash to the SHA-256 of their npm releases.