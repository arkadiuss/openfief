---
title: CI runs the full check on every push and every download retries
status: active
sourceFile: scripts/tests/ci-gate.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:37.956Z'
id: TEST-tooling-ci-gate
type: test
---
Reads the real workflows, the setup action and the scripts with gojq: ci.yml runs on pushes to main and to the implementation branch, its check job runs task check with no condition, only main publishes the docs, the images and the demo, no workflow uses the cosign installer except through the retrying install-cosign action (eight steps), the Taskfile, the download scripts and the setup action download only through retry.sh, and the tool cache is keyed by the pinned versions without the e2e state. A synthetic script with a bare curl, go install and npm ci proves the download check finds what it is for.
