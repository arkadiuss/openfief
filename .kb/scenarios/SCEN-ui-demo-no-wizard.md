---
title: A demo never shows the first-run wizard, and says when its data is loading
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T19:04:02.865Z'
id: SCEN-ui-demo-no-wizard
type: scenario
---
Given a core running with demo sessions on, when an owner session with no account opens the overview, then it stays on the overview instead of the wizard, and every step of /setup redirects to the overview; while the demo data is loading, the overview and the login page say "Demo data is loading" and the login page offers Refresh instead of View demo; once loaded, the login page offers View demo again.