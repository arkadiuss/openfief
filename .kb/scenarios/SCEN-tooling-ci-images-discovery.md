---
title: CI discovers the images to build
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:33.303Z'
id: SCEN-tooling-ci-images-discovery
type: scenario
---
Given no Dockerfile under cmd/ or plugins/, when task ci:images runs, then it prints [] and the image jobs are skipped. Given cmd/fief-operator/Dockerfile and plugins/xtb-import/Dockerfile, it lists both with lowercased registry references. A name used by two Dockerfiles, or one that is not lowercase letters, digits and dashes, fails.
