---
title: The nbp-fx connector takes an https base URL from its config, asks the grant for its host and calls it
status: active
sourceFile: plugins/nbp-fx/endpoint_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T22:47:50.069Z'
id: TEST-plugins-nbp-fx-endpoint
type: test
---
TestConfigBaseURL reads the plugin manifest.json as its signed label and tables configs: none, the default written out, a mirror with port and path, and refused shapes such as plain http, a user, a query or a fragment. Each accepted config must pass the manifest JSON Schema with the operator validator, give the config hosts plugins.ConfigHosts derives for the grant (none without a base URL, beside the fixed api.nbp.pl), and resolve with openfief.BaseURL to the configured base URL or the real default; each refused one must fail the schema.