---
title: Without a Cloudflare token, or with DEMO_DOMAIN=none, no DNS call is made
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:35:28.752Z'
id: SCEN-tooling-demo-dns-fallback
type: scenario
---
Given CLOUDFLARE_API_TOKEN is unset and DEMO_DOMAIN is openfief.dev, when task demo:up runs, then it says so, makes no Cloudflare call, and prints the URL https://demo.openfief.dev with the A record to create by hand; down says any DNS records stay. Given DEMO_DOMAIN=none, even with a token, no DNS call is made, cloud-init carries host none so the machine derives <ipv4-with-dashes>.sslip.io, and up prints that URL.