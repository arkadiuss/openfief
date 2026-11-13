---
title: The install summary of a view plugin says where to grant it accounts
status: active
sourceFile: internal/plugins/cli/grant_text_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'design-session answers for the permissions plan, 2026-11-13 (later): attention item view-plugin-ungranted (D-196, D-209)'
  recorded_at: '2026-11-13T17:06:29.915Z'
id: TEST-plugins-cli-grant-text
type: test
---
Installs a view plugin, one with a browser path and a connector against a fake cluster and checks that the output of the view plugins ends with the line about granting accounts in Settings → Access, and a connector's has none.