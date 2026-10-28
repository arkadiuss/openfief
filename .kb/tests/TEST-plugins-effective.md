---
title: Effective permissions are the intersection of the plugin's request and its grant
status: active
sourceFile: internal/plugins/effective_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:27:55.987Z'
id: TEST-plugins-effective
type: test
---
Table tests build a synthetic Plugin and Grant and call Effective. A full grant yields every requested permission, a grant approving extra hosts, Secrets, tiers and capabilities yields only the requested ones, a partial grant yields only the overlap (an approved host on another port does not count), an upgrade requesting a new host keeps only the granted ones, and a grant without the requested capability fails with ErrNoCapabilityGranted. Mismatched grants and unpinned images fail. Results are read through the accessors, and returned slices are checked to be copies. A false pass would be a union instead of an intersection; the extra-grant case catches it.