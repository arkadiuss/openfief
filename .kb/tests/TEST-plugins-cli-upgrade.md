---
title: openfief plugin upgrade shows how permissions differ and changes nothing without confirmation
status: active
sourceFile: internal/plugins/cli/upgrade_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T17:03:54.513Z'
id: TEST-plugins-cli-upgrade
type: test
verification_scope: unit
---
Runs cli.Run over the fake cluster and registry: after an install, upgrade shows both images and the added egress host, a declined answer leaves the image, a confirmed one stores the new image and exact grant, and the same version again says nothing changes. An unknown instance, a tag, a version the catalog lacks, a missing version argument and installing over the instance are refused.