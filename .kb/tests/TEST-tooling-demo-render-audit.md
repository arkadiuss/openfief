---
title: render-audit.sh renders the demo's egress audit Job at the verified egress proxy digest with the deploy's tag
status: active
sourceFile: scripts/tests/demo-render-audit.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T03:30:43.969Z'
id: TEST-tooling-demo-render-audit
type: test
---
Runs scripts/demo/render-audit.sh on a copy of deploy/demo whose egress proxy digest is pinned as pin-images.sh leaves it, with the pinned kustomize and gojq: the Job runs egress-proxy audit at that digest, once, as the egress-audit ServiceAccount in openfief, with the deploy's tag on its template and Flux's force annotation; its Role only lists pods and grants and reads pod logs; the committed directory keeps its placeholder; a placeholder or missing digest and a bad tag fail.