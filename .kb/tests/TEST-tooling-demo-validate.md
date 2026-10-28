---
title: validate.sh builds every demo layer and rejects unpinned images or an unverified OCI source
status: active
sourceFile: scripts/tests/demo-validate.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:10.339Z'
id: TEST-tooling-demo-validate
type: test
---
Runs scripts/demo/validate.sh with the pinned kustomize on copies of deploy/ in a throwaway root. The committed manifests pass; an OCI source whose verification provider is not cosign fails naming it; removing the images list of deploy/demo/openfief (images by tag only) fails naming the image; and a git-sourced Flux Kustomization path that does not exist fails naming the path. Every case also checks the message, so a failure for the wrong reason does not pass.