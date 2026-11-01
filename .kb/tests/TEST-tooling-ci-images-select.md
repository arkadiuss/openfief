---
title: CI finds each image's last signed build
status: active
sourceFile: scripts/tests/ci-images-select.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-30T03:32:36.978Z'
id: TEST-tooling-ci-images-select
type: test
verification_scope: integration
---
Runs scripts/ci-images-select.sh in a throwaway git repository with a fake cosign that verifies only listed tags for the images workflow on main: reuse mode records each image's newest signed sha-<commit> build, checking HEAD first and stopping at the first verified commit, records none for an image never signed, and hands over the diff since that build; base mode compares every image with the base and calls no cosign; all mode compares with nothing; a bad repository, revision or mode is refused.