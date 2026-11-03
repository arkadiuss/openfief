---
title: Index readings are a permission a connector declares and the grant approves
status: active
verification_scope: unit
sourceFile: internal/plugins/readings_test.go
origin:
  kind: agent
  recorded_at: '2026-11-03T10:08:02.403Z'
id: TEST-plugins-index-readings
type: test
---
Tables plugins and grants through Manifest.Validate, Plugin.Requested, NeedsApproval and Effective: the indexReadings declaration is requested, waits for approval until the grant holds it and is enforced only when both hold it, and a view can't declare it; a manifest label carries it, manifests and effective permissions differing only in it are unequal, and adding or dropping it shows in an upgrade's diff. Enforcing it without approval fails the test.