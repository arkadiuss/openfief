---
title: The cluster adapter reads an installed instance back and replaces its Plugin and grant for an upgrade
status: active
sourceFile: internal/plugins/kube/upgrade_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T17:03:52.248Z'
id: TEST-plugins-kube-upgrade
type: test
---
Over the controller-runtime fake client: Get reads back what Resources wrote, including an in-cluster service, and is ErrNotInstalled without the Plugin or the grant; Update sets the new image and permissions, takes back the operator's read access to an in-cluster service the new version no longer names, and refuses a missing instance.