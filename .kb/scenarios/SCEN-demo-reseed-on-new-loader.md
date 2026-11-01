---
title: A deploy of a new loader build reseeds the demo
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T16:48:23.761Z'
id: SCEN-demo-reseed-on-new-loader
type: scenario
---
Given a demo database marked with the build of the loader that filled it, when another build of the loader runs without --reset-demo, then it empties the demo's data and loads the dataset again, recording its own build; the same build loads again without change, and a database without the marker is refused.