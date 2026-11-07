---
title: The treasury-bonds connector takes an https base URL from its config, asks the grant for its host and calls it
status: active
sourceFile: plugins/treasury-bonds/endpoint_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T22:47:51.966Z'
id: TEST-plugins-treasury-bonds-endpoint
type: test
---
TestConfigBaseURLs reads the plugin manifest.json as its signed label and tables configs: none, the default written out, a mirror with port and path, and refused shapes such as plain http, a user, a query or a fragment. Each accepted config must pass the manifest JSON Schema with the operator validator, give the config hosts plugins.ConfigHosts derives for the grant (none without a base URL, beside the fixed api-sdp.stat.gov.pl and static.nbp.pl), and resolve with openfief.BaseURL to the configured base URL or the real default; each refused one must fail the schema. Both gusBaseUrl and nbpBaseUrl are checked, alone and together, and newConnector is built with clients at the configured base URLs.