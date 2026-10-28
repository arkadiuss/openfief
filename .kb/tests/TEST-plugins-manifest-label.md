---
title: The image manifest label is parsed strictly and compared as sets
status: active
sourceFile: internal/plugins/manifest_label_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:50:29.714Z'
id: TEST-plugins-manifest-label
type: test
---
Table tests parse manifest labels with every field, reject a label without a capability, with an unknown capability, an IP egress host, a reserved mount path, an unknown field or non-JSON text, and compare a declared manifest with a parsed label: lists in another order and a reformatted config schema are equal, while an extra egress host, another port, an extra capability, another mount path, an extra data tier and another or missing schema are not. A false pass would be order-sensitive or byte-wise comparison; the reordered label covers both.