---
title: An IKE takes the earlier payments the owner enters
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:20.147Z'
id: SCEN-ingest-earlier-payments-checks
type: scenario
---
Given an IKE, when the owner enters 40000 PLN as its earlier payments, then they are stored with the account, served with it and shown in its change form; earlier payments for a wrapper that takes none, below zero or in another currency are refused before anything is stored.