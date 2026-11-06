---
title: Completed import batches page newest first
status: active
origin:
  kind: agent
  recorded_at: '2026-11-05T23:39:01.906Z'
id: SCEN-ingest-completed-batches-pages
type: scenario
---
Given completed sync batches of two instances and an open one, when the completed batches are listed, then they come newest first by when they finished with their counts, the open one is left out, and passing the last one as the cursor returns the rest.