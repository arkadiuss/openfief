---
title: A reusable signing workflow refuses to sign for another repository
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T02:14:25.170Z'
id: SCEN-tooling-reusable-signing-guard
type: scenario
---
Given images.yml and deploy-demo.yml, which other workflows call and which sign keylessly, when a workflow of another repository calls either of them, then github.repository is that caller's repository, the signing job's if condition is false and the job is skipped, so no OIDC token is requested and no certificate naming the workflow is issued for the caller's build; when OpenFief's own ci.yml or release.yml calls them, the signing jobs run as before. Any workflow with a workflow_call trigger and a job with id-token: write must carry the same condition.