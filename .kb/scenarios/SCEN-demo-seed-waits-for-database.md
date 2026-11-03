---
title: The demo loader waits for a database that is still starting
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T17:14:57.284Z'
id: SCEN-demo-seed-waits-for-database
type: scenario
---
Given a database that refuses the first connections, when the demo loader starts, then it retries with a backoff that doubles from 0.5 s up to 5 s and loads once the database accepts connections; a database still refusing after --db-wait fails the attempt with the last error.