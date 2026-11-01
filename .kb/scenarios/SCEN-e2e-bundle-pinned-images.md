---
title: An install from the bundle runs its images by digest
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:32.768Z'
id: SCEN-e2e-bundle-pinned-images
type: scenario
---
Given the release bundle rendered from this checkout's images in the suite's registry and applied unchanged to a kind or k3d cluster, when the suite reads the bundle's Deployments and the live ones, then every container image is the suite's registry at a sha256 digest, the cluster runs exactly the images the bundle names, and the bundle holds exactly the operator, the egress proxy and the core.