---
title: A demo core reconciles SpiceDB as soon as the demo's loader changes its accounts, and every five minutes
status: active
origin:
  kind: agent
  recorded_at: '2026-11-22T08:57:45.725Z'
id: SCEN-demo-spicedb-reconciled-after-reseed
type: scenario
---
Given a core started with --demo-sessions under SpiceDB, when its start's reconcile is still retrying, then it reconciles nothing more; once that has succeeded, then it reads the accounts and their sleeves every poll (15 seconds in production) and reconciles SpiceDB the first time and whenever they differ from those it last reconciled, in any order, so accounts the demo's loader wrote after the start get their household relationship and demo sessions view them within about one poll; it reconciles nothing while the ledger is empty, between the loader's drop and its accounts; a failed reconcile is logged as a warning and retried at the next poll; every sweep (five minutes) it reconciles whatever it read; and it stops when the core shuts down. Assumes the demo's account IDs stay the same across reseeds, as the synthetic dataset fixes them, and that the loader writes every account in its first step.