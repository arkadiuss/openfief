---
title: The SDK takes the core address, token path, listen port and config from the operator's environment
status: active
sourceFile: sdk/go/openfief/env_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:36.509Z'
id: TEST-sdk-go-env
type: test
verification_scope: unit
---
Sets the process environment the way the operator does, with synthetic values: CoreURL turns OPENFIEF_CORE_ADDRESS 10.96.0.10:8080 into http://10.96.0.10:8080 and refuses an empty address, a DNS name without a port and a URL with a scheme. ListenAddress turns port 9090 into :9090 and refuses empty, zero, non-numeric and out-of-range ports. LoadConfig decodes OPENFIEF_CONFIG JSON, leaves the value unchanged without config and fails on malformed JSON. A zero TokenFile reads the file OPENFIEF_TOKEN_FILE names.