---
title: The core and its tests import no UI package, and another binary may
status: active
origin:
  kind: agent
  ref: Design session decision on the MCP toolbox seam, 2026-11-18
  recorded_at: '2026-11-19T11:06:19.614Z'
id: SCEN-tooling-archtest-core-no-ui
type: scenario
---
The architecture test loads a repository where cmd/core, a package below it and a test of the core each import internal/views/ui, and another binary does too. It reports the three imports of the core and not the other binary, so the core links none of the UI package and its image needs no design system.
