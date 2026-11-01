---
title: openfief plugin install and upgrade refuse a second view with a browser path before the owner confirms
status: active
sourceFile: internal/plugins/cli/browser_view_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T21:58:45.998Z'
id: TEST-plugins-cli-browser-view
type: test
verification_scope: unit
---
Runs the command line over the in-memory cluster with the instance ui holding browser path /: installing a signed view that declares /gallery fails with ErrBrowserPathHeld even with the owner answering yes, and installs nothing; upgrading a plugin without a browser path to that view fails the same way and keeps its image; on a cluster with no view, the same install goes through.