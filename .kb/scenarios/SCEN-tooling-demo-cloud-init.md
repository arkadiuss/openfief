---
title: cloud-init is rendered only with valid pins
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:39.873Z'
id: SCEN-tooling-demo-cloud-init
type: scenario
---
Given the pinned k3s and Flux versions and checksums, when cloud-init is rendered for a cax11, then it downloads k3s-arm64 checked against the arm64 sha256, Flux's install.yaml of the pinned version checked against its sha256, disables Traefik and servicelb, and leaves no placeholder. A missing checksum, a version such as latest, a value with a newline or an unknown placeholder makes rendering fail.