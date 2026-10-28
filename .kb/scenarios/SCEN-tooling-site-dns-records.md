---
title: task dns:site points the apex and www at GitHub Pages
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T15:29:24.830Z'
id: SCEN-tooling-site-dns-records
type: scenario
---
Given CLOUDFLARE_API_TOKEN for the zone example.com and no records, when task dns:site:plan runs, then it shows A records example.com to 185.199.108.153, .109, .110 and .111, AAAA records to 2606:50c0:8000::153 through 2606:50c0:8003::153 and a CNAME www.example.com to arkadiuss.github.io, and changes nothing. task dns:site creates exactly those nine records, DNS-only, each with the comment managed-by=openfief-site, and a second run changes nothing. When one of our A records holds an address no longer listed, another extra address of ours exists and one AAAA is proxied, then the missing address is created first, both stale ones are deleted and the proxied one is set DNS-only. Without the token or with DEMO_DOMAIN=none it fails without an API call.