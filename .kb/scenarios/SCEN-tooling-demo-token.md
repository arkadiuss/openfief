---
title: The API tokens are required where needed and never shown
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:34.825Z'
id: SCEN-tooling-demo-token
type: scenario
---
Given neither HCLOUD_TOKEN nor HETZNER_API_TOKEN is set, when any demo command runs, then it fails naming HCLOUD_TOKEN without calling the API; HETZNER_API_TOKEN alone is accepted. Given the Hetzner token and CLOUDFLARE_API_TOKEN, then no command line, output or written file of plan, up or down contains either; the Cloudflare token reaches curl on standard input only, and a rejected token shows only Cloudflare's error.