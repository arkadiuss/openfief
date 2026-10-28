---
title: The CLI shows an install's trust and every requested permission, an uninstall's deletions, and the installed instances
status: active
sourceFile: internal/plugins/cli/render_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:35:55.882Z'
id: TEST-plugins-cli-render
type: test
---
Golden-text tests in package cli_test render a verified and an unverified synthetic proposal with two egress hosts, a Secret and its mount path, a data tier and a config schema whose properties have a type, a type list, an enum, no type and a required flag, and one with no egress, Secrets or tiers and a schema without properties. They render an uninstall plan that deletes two Secrets and keeps a shared one, and the list table. A false pass would be a summary that drops a permission; every requested item is in the golden text.