---
title: validate.sh builds every demo layer and rejects unpinned images, unverified OCI sources or unknown variables
status: active
sourceFile: scripts/tests/demo-validate.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:10.339Z'
id: TEST-tooling-demo-validate
type: test
verification_scope: integration
---
Runs scripts/demo/validate.sh with the pinned kustomize on copies of deploy/ in a throwaway root. The committed manifests pass; an OCI source whose verification provider is not cosign fails naming it; removing the images list of deploy/demo/openfief (images by tag only) fails naming the image; a git-sourced Flux Kustomization path that does not exist fails naming the path; and a substituted layer using a variable the demo settings do not define fails naming it. Every case also checks the message, so a failure for the wrong reason does not pass.