---
title: The check of view plugins without a grant
status: active
sourceFile: cmd/core/viewgrants_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'design-session answers for the permissions plan, 2026-11-13 (later): attention item view-plugin-ungranted (D-196, D-209)'
  recorded_at: '2026-11-13T17:06:29.216Z'
id: TEST-core-view-grants
type: test
---
Checks that under SpiceDB each Ready view plugin without a grant in force is a condition naming its instance, that a grant of any target, a plugin not Ready or being removed, one without the view capability, the plugin serving the browser path, the policy and no cluster each find nothing, that a plugin cache not synced or unreadable grants fail the check, and that no cluster is a nil port.