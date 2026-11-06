---
title: No exported span of the core's or the SDK's traffic carries money, holdings, accounts or credentials
status: active
sourceFile: cmd/core/tracing_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-02T03:05:47.805Z'
id: TEST-core-tracing-privacy
type: test
---
Drives representative traffic through the production instrumentation into tracetest's in-memory exporter behind telemetry.Redact: two syncs from the core's connector client (pluginClient) to a connector served by the SDK's NewConnectorServer, the second refused with a message quoting an account, ISIN and quantity; a plugin's call to the core's plugin port through CoreHTTPClient and a browser's call to its browser port with a session cookie, a bearer token, an account number, an ISIN and an IBAN in the path and a forged traceparent; a query and a failure quoting an amount through queryTracer, and a real query through newPool when task test:pg gives a database; a revaluation job enqueued and worked through jobTracing. It checks the plugin ran in the core's trace, the work span links to its insert, the browser's forged trace is not continued, and every exported name, attribute, event, link and status against a deny-list of keys and patterns for amounts, ISINs, IBANs, JWTs and the synthetic values. withTracing with no endpoint records nothing. Removing Redact makes it fail.