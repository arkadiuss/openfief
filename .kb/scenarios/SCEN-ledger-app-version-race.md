---
title: Two imports of one correction at once append one version
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:24.733Z'
id: SCEN-ledger-app-version-race
type: scenario
---
Given a transaction stored as version 1, when two or more imports of the same corrected row run at the same time and each decides that version 2 is next, then the store accepts only the first append of version 2, and the others reload, find version 2 unchanged and append nothing. When another import keeps winning, the import stops after a bounded number of attempts with a version conflict error, and running it again is safe.