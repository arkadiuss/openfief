---
title: The export filter keeps only allowed attributes and no error message
status: active
sourceFile: sdk/go/telemetry/redact_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T03:05:43.570Z'
id: TEST-telemetry-redact
type: test
---
A span stub carrying HTTP attributes, a connector instance, a URL with a token, a SQL statement with parameters, a client address, an amount attribute, an exception event with a message quoting an account, a link attribute with a session and an error status quoting a price passes through telemetry.Redact into tracetest's in-memory exporter. The exported span keeps its name, the method, the status code and the instance, the exception's type alone, the link without attributes, and the error code without its description.