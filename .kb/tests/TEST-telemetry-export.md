---
title: Spans reach an OTLP collector only when flushed, redacted, with the trace propagated
status: active
sourceFile: sdk/go/telemetry/telemetry_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-02T03:05:45.680Z'
id: TEST-telemetry-export
type: test
---
Starts tracing with no endpoint and checks nothing records; then points it at an httptest OTLP/HTTP receiver that decodes ExportTraceServiceRequest. Within a root span it calls a ServerHandler, which continues the caller's trace, a PublicHandler, which starts a new one, and a health probe, which gets no span, through telemetry.Transport, and runs a failing telemetry.Run with an amount attribute. Nothing arrives before Shutdown; after it the receiver holds every span, named by procedure or method, the Run span a child of the root, an error without a message, with the instance alone. A gRPC exporter starts and shuts down too.