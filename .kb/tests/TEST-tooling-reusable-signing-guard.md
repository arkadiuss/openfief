---
title: Every reusable workflow's signing jobs run only in OpenFief's own repository
status: active
sourceFile: scripts/tests/reusable-signing-guard.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T02:14:17.725Z'
id: TEST-tooling-reusable-signing-guard
type: test
---
Parses every workflow under .github/workflows with gojq and fails when a workflow with a workflow_call trigger has a job that may request an OIDC token, through id-token: write, write-all or the workflow's own permissions, whose if condition doesn't require github.repository == 'arkadiuss/openfief' as a top-level conjunct. It also checks that the reusable signing workflows are exactly images.yml and deploy-demo.yml, so the check can't pass by seeing nothing, and runs synthetic workflows the check must refuse: no guard, the guard inside an or, another repository, and a job inheriting the workflow's permission; guarded jobs, jobs that don't sign and workflows only this repository triggers pass. A false pass would need GitHub to evaluate if differently from its text.