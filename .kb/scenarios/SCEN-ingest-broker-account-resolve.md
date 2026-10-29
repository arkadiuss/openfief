---
title: Entries of a mapped broker account are booked in its account; those of an unmapped one are not
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:56.196Z'
id: SCEN-ingest-broker-account-resolve
type: scenario
---
Given U0000001 mapped to acc-1 and U0000002 unmapped, when a delivery names entries of both by broker account with the key source test, then U0000001's are booked in acc-1 with source plugin:ibkr-main:test:acc-1, and nothing of U0000002's is booked; an IBKR statement resolved this way books exactly what the mapper configured with that account booked.