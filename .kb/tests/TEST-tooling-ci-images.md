---
title: ci-images.sh lists Dockerfiles as a build matrix
status: active
sourceFile: scripts/tests/ci-images.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:32.146Z'
id: TEST-tooling-ci-images
type: test
verification_scope: integration
---
Runs scripts/ci-images.sh in a throwaway directory: no Dockerfile gives [], two Dockerfiles give two matrix entries with lowercased image references, and a duplicate or invalid image name fails with a message naming it.
