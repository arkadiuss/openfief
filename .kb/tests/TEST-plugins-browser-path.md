---
title: A view plugin's browser path is validated, requested, approved by its grant and made effective
status: active
sourceFile: internal/plugins/browser_path_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:37.452Z'
id: TEST-plugins-browser-path
type: test
verification_scope: unit
---
Table tests over synthetic view plugins and grants: CoreServesPath keeps /healthz, /readyz, /v1 and the core's Connect service paths for the core and nothing else; a manifest's browser path must be / or a clean lower-case path the core doesn't serve, and only a view may declare one; Requested lists it, NeedsApproval reports it until the grant approves it, Effective keeps it only when approved and when view is effective, and the manifest label parses and compares it. A pass with a broken check would need a refused path to validate.