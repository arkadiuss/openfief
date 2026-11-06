---
title: The core refuses the wizard's writes while a demo's data loads
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T19:04:00.096Z'
id: SCEN-identity-demo-loading-refuses-wizard-writes
type: scenario
---
Given a core running with demo sessions on and a database without the demo marker yet, when the login page asks for the login methods, then the core reports no demo sessions and the demo data as loading; and when the UI hands it a sleeve or an account to store, the core refuses with FAILED_PRECONDITION and stores nothing, while reads are served. Once the marker exists, it reports demo sessions and accepts the owner's changes again; without demo sessions nothing changes.