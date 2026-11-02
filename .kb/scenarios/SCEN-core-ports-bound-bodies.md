---
title: The core's ports cap request bodies by path and refuse compressed ones
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:15.120Z'
id: SCEN-core-ports-bound-bodies
type: scenario
---
Given the core's browser and plugin ports, when a request to the core's own APIs carries more than 2 MiB, it is cut off with 413 before the API decodes it; when the browser sends a statement form to the view plugin or a view plugin streams a statement upload, up to 65 MiB passes; a compressed body is refused with 415; and both servers bound how long a request may take to arrive and a connection may idle.