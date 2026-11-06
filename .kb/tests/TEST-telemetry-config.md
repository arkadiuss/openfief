---
title: 'Tracing configuration: off by default, the standard variables, bad values refused'
status: active
sourceFile: sdk/go/telemetry/config_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T03:05:41.485Z'
id: TEST-telemetry-config
type: test
---
Table tests of telemetry.ParseConfig over a fake environment: no variables leave export off; a base endpoint over http/protobuf gets /v1/traces and a traces endpoint is used as given and wins; grpc keeps the endpoint; the traces protocol wins; service name, ratio sampler and resource attributes are read; OTEL_SDK_DISABLED and the none exporter turn an endpoint off; headers are URL-decoded. Every bad value (scheme, protocol, headers, sampler, ratio, exporter, boolean, resource attributes) is refused, also while export is off, and the error never quotes a header. TestConfigSampler checks each sampler name builds the sampler it names.