---
title: The demo's model is ready and off, and the seed never turns it on
status: active
origin:
  kind: agent
  ref: 'M4 prep brief from the design session: the demo''s model behind the switch (D-207 amended, D-205, D-208, D-212, D-203), 2026-11-14'
  recorded_at: '2026-11-14T17:59:32.838Z'
id: SCEN-demo-model-prepared-off
type: scenario
---
Given the demo's model layer, when the demo is validated, then the layer builds with the demo, its images are pinned by digest, the live demo's manifests do not include it, the settings route only the Ask box and name only entries the instance lists; when the demo loader runs with the settings file, then an empty modelrouter schema gets the route and the tier lists, the switch and every toggle stay off, and a database that already has a route or a tier list keeps it.