---
title: The ibkr-flex connector takes an https base URL from its config, asks the grant for its host and calls it
status: active
sourceFile: plugins/ibkr-flex/endpoint_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T22:47:51.345Z'
id: TEST-plugins-ibkr-flex-endpoint
type: test
---
TestConfigBaseURL reads the plugin manifest.json as its signed label and tables configs: none, the default written out, a mirror with port and path, and refused shapes such as plain http, a user, a query or a fragment. Each accepted config must pass the manifest JSON Schema with the operator validator, give the config hosts plugins.ConfigHosts derives for the grant (none without a base URL, beside the fixed ndcdyn.interactivebrokers.com), and resolve with openfief.BaseURL to the configured base URL or the real default; each refused one must fail the schema. TestFlexEndpoints checks that the default base URL gives exactly the SendRequest and GetStatement endpoints IBKR documents and that a mirror gives them below its own base URL.