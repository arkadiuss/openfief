---
title: Images built or carried for each kind of change
status: active
sourceFile: tools/affected/images_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T03:32:22.018Z'
id: TEST-tooling-affected-images
type: test
verification_scope: unit
---
Table tests decide for a command, a plugin on the SDK and an image embedding a file whether each is built or keeps its earlier signed digest: docs and kibi changes and test files carry every image; a package in one binary's import closure rebuilds that image only; a shared package rebuilds every importer; an SDK change rebuilds the plugin; an embedded file and a Dockerfile rebuild their image; go.sum, the images workflow and a Go file in no package rebuild everything; no earlier build or no diff builds. Each carry is checked to keep the earlier build's digest and commit, so a carry of the wrong digest fails.