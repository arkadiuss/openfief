---
title: A grant approves a connector's institutions like its hostnames
status: active
sourceFile: internal/plugins/grant_institutions_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T07:22:09.291Z'
id: TEST-plugins-grant-institutions
type: test
---
Table tests over the sample connector and a grant of everything else: NeedsApproval lists the declared institutions the grant doesn't approve, Effective keeps only those both name and none the grant alone names, a grant's institutions are validated like a manifest's, DiffPermissions names the institutions an upgrade adds, ApprovedInstitutions returns a grant's own list, an empty one included, or for a grant without one the running version's, and SharedInstitutions names each institution another instance books for and never the instance itself.