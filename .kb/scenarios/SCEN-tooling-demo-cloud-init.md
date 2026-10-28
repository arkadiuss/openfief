---
title: cloud-init is rendered only with valid pins and settings
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:39.873Z'
id: SCEN-tooling-demo-cloud-init
type: scenario
---
Given the pinned k3s and Flux versions and checksums, when cloud-init is rendered for a cax11, then it downloads k3s-arm64 checked against the arm64 sha256 and Flux's install.yaml of the pinned version checked against its sha256, keeps Traefik and servicelb, writes DEMO_HOST from the machine's IPv4 as an sslip.io name and DEMO_ACME_SERVER for Let's Encrypt production, or staging with DEMO_ACME=staging, and leaves no placeholder. A missing checksum, a version such as latest, an unknown DEMO_ACME, a value with a newline or an unknown placeholder makes rendering fail.