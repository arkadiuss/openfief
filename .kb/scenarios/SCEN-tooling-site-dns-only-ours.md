---
title: task dns:site touches only its records and the retired docs record
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T15:29:28.388Z'
id: SCEN-tooling-site-dns-only-ours
type: scenario
---
Given a foreign A record at example.com, a foreign CNAME at the apex, a foreign AAAA at the apex or any foreign record at www.example.com, when task dns:site runs, then it refuses naming the record and changes nothing, ours included; an MX record at the apex is kept. Given a CNAME docs.example.com with the retired comment managed-by=openfief-docs, then plan shows its deletion and apply deletes it and nothing else, while a docs record without that comment is left alone and no docs record is ever created. In a zone holding the demo's records, dns:site leaves them unchanged; the demo's down keeps every site record, and a demo named www is refused at the site's www record.