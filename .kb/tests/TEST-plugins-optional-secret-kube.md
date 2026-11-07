---
title: An optional Secret stays optional through the Plugin resource the CLI writes and reads back
status: active
sourceFile: internal/plugins/kube/optional_secret_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T21:54:59.048Z'
id: TEST-plugins-optional-secret-kube
type: test
---
TestInstancesKeepOptionalSecret plans an install with an optional ca-bundle Secret and a required credentials Secret, writes the Plugin and grant with kube.Resources into a fake API server and reads the instance back with Instances.Get, as an upgrade does. The read-back manifest Secrets must equal the planned ones, optional flag included; losing the flag would make the upgrade require a Secret the owner never had to create.