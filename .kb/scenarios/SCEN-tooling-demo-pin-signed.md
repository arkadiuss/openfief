---
title: The demo deploys only verified digests
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:45.636Z'
id: SCEN-tooling-demo-pin-signed
type: scenario
---
Given the operator and egress-proxy images tagged sha-1a2b3c4, when task deploy:demo runs in CI after a push to main and both signatures from the images workflow on main verify, then both digests are written into CI's copy of deploy/demo/openfief, which is rendered, published and signed as the artifact Flux follows; when one does not verify, the run fails before anything is published.