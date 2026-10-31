---
title: The Ghostfolio plugin asks only for the view capability and its token Secret, checks its config and syncs on a timer
status: active
sourceFile: plugins/ghostfolio/main_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:36:51.736Z'
id: TEST-plugins-ghostfolio-main
type: test
verification_scope: unit
---
TestManifest parses manifest.json with ParseManifestLabel and requires exactly the view capability, the ghostfolio-credentials Secret at /etc/openfief/ghostfolio, no egress host and a config schema. TestConfigParse accepts an https origin with an optional interval and listings, and refuses plain http, a path, a user in the URL, an interval of one minute or not a duration, a wrong ISIN check digit and an ISIN twice. TestReadSecretFile reads and trims the token and refuses a missing, empty or escaping file. TestNewSyncer refuses to start without the core's address, the instance name or the egress proxy; TestNewSyncerWith that a missing token fails the sync before any request; TestLoop that the loop syncs at once, goes on after a failure and stops with its context.