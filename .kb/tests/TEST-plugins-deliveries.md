---
title: Bond terms and index levels are permissions a connector declares and the grant approves
status: active
verification_scope: unit
sourceFile: internal/plugins/deliveries_test.go
origin:
  kind: agent
  recorded_at: '2026-11-02T22:24:38.513Z'
id: TEST-plugins-deliveries
type: test
---
Tables plugins and grants through Manifest.Validate, Plugin.Requested, NeedsApproval and Effective: each declaration is requested, waits for approval until the grant holds it and is enforced only when both hold it, a view can't declare either, a manifest label carries both, and adding or dropping one shows in an upgrade's permissions diff. A permission granted without the declaration, or enforced without approval, fails it.