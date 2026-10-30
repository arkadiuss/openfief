---
title: Only affected images are rebuilt
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T03:32:59.042Z'
id: SCEN-tooling-ci-affected-images
type: scenario
---
Given a push to main that changes a package one plugin imports, when CI decides the images, then that plugin's image is built, signed and published, the images whose import closure, Dockerfile and shared inputs did not change since their last signed build are not rebuilt, and the job log names each decision and its reason.