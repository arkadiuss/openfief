---
title: task check validates the demo manifests
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:51.777Z'
id: SCEN-tooling-demo-manifests-check
type: scenario
---
Given deploy/demo, when task demo:validate runs, then deploy/demo, every git path its Flux Kustomizations apply and the deploy/demo/openfief template build with the pinned kustomize, every image is pinned by a sha256 digest, and every OCI source requires a cosign signature; otherwise it fails naming the problem.