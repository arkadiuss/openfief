---
title: The demo looks no quote up, whatever connector is installed
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: SCEN-demo-market-quotes-off
type: scenario
---
Given the core started with demo sessions and a market-data connector installed and approved, when the views are asked for a quote or for the settings, then the lookup is a failed precondition before any connector is asked and the settings say market quotes are not available, so the Ask box does not offer the tool. Assumes the demo flag is the one the demo's core is started with.
