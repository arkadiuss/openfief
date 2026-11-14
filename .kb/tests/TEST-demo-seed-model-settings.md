---
title: The demo loader stores the demo model's route and tier lists once, never turns a model on and leaves the owner's settings alone
status: active
sourceFile: cmd/demo-seed/modelsettings_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'M4 prep brief from the design session: the demo''s model behind the switch (D-207 amended, D-205, D-208, D-212, D-203), 2026-11-14'
  recorded_at: '2026-11-14T17:58:43.058Z'
id: TEST-demo-seed-model-settings
type: test
---
Table tests parse the settings file the way the router would check the owner's own: known roles and tiers, provider/model references, a route of distinct models; a member for the switch or a toggle, an unknown member and bad references are refused. A fake store shows the seed stores the route and the tier lists in an empty store (and one with only the switch saved as off), and nothing in a store that has a route or a tier list. The committed deploy/demo/model/settings.json parses, routes only the Ask box to the demo's model and allows it for every tier. With OPENFIEF_PG_DSN set, a real database gets the route and the tiers with the switch off and no toggle on, no maintainer route, and a second run after the owner changed the route keeps the owner's.