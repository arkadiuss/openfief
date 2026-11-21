---
title: Who may read market quotes
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md), and the design session's ruling on it
  recorded_at: '2026-11-21T14:00:00.000Z'
id: SCEN-views-market-quote-callers
type: scenario
---
Given the read API with a quote source, when the owner, a script token or a viewer asks for a quote then it is answered; when a view plugin or a demo session asks, then it is permission denied and the source is not asked, and the settings say quotes are not available to them.
