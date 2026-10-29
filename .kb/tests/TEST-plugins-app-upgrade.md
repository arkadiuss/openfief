---
title: PrepareUpgrade verifies the new image like an install and Upgrade stores the new Plugin and its exact grant
status: active
sourceFile: internal/plugins/app/upgrade_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T17:03:50.067Z'
id: TEST-plugins-app-upgrade
type: test
---
Over the app package's fake registry and in-memory store: upgrading an installed sample to another signed image returns the installed and new version, an empty diff and unverified trust; Upgrade stores the new image and a grant equal to what it requests. An unknown instance, an untrusted image and a config the schema refuses fail.