---
title: Duplicate triples in a published store are removed and reported
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T19:27:55.349Z'
id: SCEN-tooling-kibi-duplicates-reported
type: scenario
---
Given a published kibi store that holds some entities and relationships twice, when the kibi engine attaches it, then one copy of each triple stays and the number of duplicate triples and of entities stored twice is recorded and printed by the next kibi command, as a workflow warning under GitHub Actions; a store without duplicates is left alone and nothing is printed.