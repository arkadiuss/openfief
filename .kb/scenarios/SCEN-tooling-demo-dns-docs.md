---
title: task dns:docs keeps the docs site's CNAME and nothing else
status: closed
origin:
  kind: agent
  recorded_at: '2026-10-28T14:35:30.833Z'
id: SCEN-tooling-demo-dns-docs
type: scenario
---
Given CLOUDFLARE_API_TOKEN for example.com, when task dns:docs:plan runs on an empty zone, then it shows a CNAME docs.example.com -> arkadiuss.github.io and changes nothing; task dns:docs creates it DNS-only with the comment managed-by=openfief-docs, a second run changes nothing, and a record of ours with an old target is updated in place. A record without that comment at docs.example.com is refused and left alone; the demo's records are never touched, the demo's down keeps the docs record, and a demo named docs is refused at that name. Without the token or with DEMO_DOMAIN=none it fails without an API call.