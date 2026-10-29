---
title: Install arguments resolve to a catalog entry or a pinned image, and installs grant exactly what is requested
status: active
sourceFile: internal/plugins/install_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T22:23:46.192Z'
id: TEST-plugins-install
type: test
---
Table tests in package plugins_test resolve install arguments against a synthetic catalog: a catalog name, a mirror of a pinned digest (found by digest), an image the catalog does not pin, a tag (refused as not pinned), an unknown name and a name two entries share (refused as ambiguous), and check the default instance name. PlanInstall is checked to request exactly the signed manifest and to grant exactly its requested permissions, so NeedsApproval leaves nothing pending, and to refuse names the CRD would refuse. SecretsToDelete is checked to keep Secrets another instance still names. A false pass would be a grant wider than the request; the NeedsApproval and grant comparisons catch it.