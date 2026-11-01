---
title: Only a declared and approved market-data connector is asked to price what is held
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T10:06:53.187Z'
id: SCEN-ingest-market-data-scope-declared
type: scenario
---
Given a held instrument, a treasury bond with stored terms and a model portfolio that is on, when the scheduler syncs an instance whose manifest declares a market-data connector and whose grant approves it, bound to accounts or not, then its request names the held instrument and the model's targets once each and leaves the bond out; an instance bound to no account that isn't declared and approved, such as nbp-fx, and a broker connector are asked for nothing (D-93, D-123).