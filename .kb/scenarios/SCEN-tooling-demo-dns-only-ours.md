---
title: DNS changes touch only records with the demo's comment
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:35:26.378Z'
id: SCEN-tooling-demo-dns-only-ours
type: scenario
---
Given a zone holding our records, another demo's record, an A record of www and a CNAME for docs without a comment, and a TXT record at the demo's name, when up or down runs for ENV=demo, then only records with the comment managed-by=openfief-demo env=demo are created, updated or deleted, and every other record is left as it was. Given an A record at demo.example.com without the comment, then up refuses before creating anything and leaves it, and down deletes nothing.