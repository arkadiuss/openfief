---
title: SCEN-models-entry-data-collection
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T19:04:58.600Z'
id: SCEN-models-entry-data-collection
type: scenario
---
Given an instance at openrouter.ai with one entry pinned to a host that denies data collection and another that denies it and nothing else, when each sends a request with or without tools, then provider holds data_collection deny beside the entry's order, allow_fallbacks false and require_parameters, or alone, for that entry alone; an entry with none sends what it always did, the shared client is not changed, and a config that sets dataCollection on any other base URL or to a value other than allow or deny is refused, as the manifest's schema refuses the value; the demo's model entry sets deny.