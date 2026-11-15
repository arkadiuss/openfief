---
title: The UI's config sets the Ask box's limits within a maximum each
status: active
sourceFile: plugins/ui/config_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-14T20:37:56.473Z'
id: TEST-plugins-ui-config
type: test
---
Table tests of the plugin's config: no config leaves D-208's defaults, each limit and the deadline are read, out-of-range values and a deadline that is no duration are refused, and the defaults are D-208's.