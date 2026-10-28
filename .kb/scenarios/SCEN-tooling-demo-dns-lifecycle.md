---
title: The demo's DNS records follow its server
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:35:24.219Z'
id: SCEN-tooling-demo-dns-lifecycle
type: scenario
---
Given CLOUDFLARE_API_TOKEN for the zone example.com and no records, when task demo:up ENV=demo creates the server with IPv4 192.0.2.10, then a DNS-only A record demo.example.com -> 192.0.2.10 and a CAA record 0 issue letsencrypt.org are created after the server, both with the comment managed-by=openfief-demo env=demo, and cloud-init carries demo.example.com. A second up changes nothing; an up after the IPv4 became 192.0.2.20 updates the A record in place; recreate keeps the records and updates them. The demo try-beta lives at try-beta.example.com, as every demo lives at <ENV>.<DEMO_DOMAIN>.