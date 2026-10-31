---
title: An upgrade's version resolves among the instance's plugin's catalog entries and keeps the instance
status: active
sourceFile: internal/plugins/upgrade_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T17:03:47.935Z'
id: TEST-plugins-upgrade
type: test
verification_scope: unit
---
Over a synthetic catalog of three versions of one plugin and another plugin of the same version number: a version resolves to that plugin's entry, found by the current digest, by a mirror of it or by the repository; an image reference is read as for an install; a missing version is ErrNoUpgradeVersion and a tag is refused. A shared catalog name lists every version. PlanUpgrade keeps the name, config and in-cluster service and DiffPermissions reports the added and removed hosts.