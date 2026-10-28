---
title: Plugin and Grant validation, and the exact list of permissions an upgrade still needs approved
status: active
sourceFile: internal/plugins/permissions_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:07.218Z'
id: TEST-plugins-plugin-grant
type: test
---
Table tests validate a synthetic Plugin (tagged image, bad names, manifest errors prefixed with manifest., config that is not JSON) and Grant (bad plugin name, unknown capability, IP literal, duplicate Secret, bad data tier). NeedsApproval is called on a full grant with upgrades that keep, shrink or move permissions (nothing to approve, including a new image digest and a moved Secret mount) and with upgrades that add an egress host, a port on a granted host, a Secret, a data tier or a capability, and the returned permissions must list exactly those additions, sorted. Mismatched, invalid and nil inputs return their sentinel errors. A false pass would be a diff that ignores ports or Secrets; both have cases.