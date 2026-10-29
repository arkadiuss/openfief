---
title: A broker account mapped after its instance was seeded is seeded by the next sync
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:58.876Z'
id: SCEN-ingest-broker-account-seeding
type: scenario
---
Given an instance whose opening batch completed while nothing was mapped, when the owner maps its broker account, then the instance's next claim is not seeded, so the next sync is an opening batch; that batch seeds the account, which is then seeded, and a later opening batch leaves its opening lots and opening cash out but keeps its balance assertions.