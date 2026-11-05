---
title: The built-in UI must never show the first-run wizard on a demo deployment, redirecting its pages to the overview, and while the demo data is loading its login page must say so and ask the visitor to refresh in a minute instead of offering View demo.
status: open
text_ref: docs/dev/demo.md#demo-data
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-demo-data
  - lane:strict
semantic_text: The built-in UI must never show the first-run wizard on a demo deployment, redirecting its pages to the overview, and while the demo data is loading its login page must say so and ask the visitor to refresh in a minute instead of offering View demo.
logic_claims:
  - CLAIM-456F7D09E4BA8211
semantic_clauses:
  - The built-in UI must never show the first-run wizard on a demo deployment, redirecting its pages to the overview, and while the demo data is loading its login page must say so and ask the visitor to refresh in a minute instead of offering View demo.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b9dc8ed51c114b25da285b44b31c280cc55cbb11b7280b70359f767c0b4854fa
semantic_inventory:
  - claim_key: CLAIM-456F7D09E4BA8211
    claim_text: The built-in UI must never show the first-run wizard on a demo deployment, redirecting its pages to the overview, and while the demo data is loading its login page must say so and ask the visitor to refresh in a minute instead of offering View demo
    role: normative
    status: modeled
    span:
      start: 0
      end: 248
origin:
  kind: agent
  recorded_at: '2026-11-03T19:04:25.491Z'
id: REQ-ui-demo-no-wizard
type: req
---
The built-in UI must never show the first-run wizard on a demo deployment, redirecting its pages to the overview, and while the demo data is loading its login page must say so and ask the visitor to refresh in a minute instead of offering View demo.

## Context

On a fresh demo database, before cmd/demo-seed has written the demo marker, a visitor who enrolled a passkey landed in the first-run wizard, whose sleeves and accounts are data without the marker; the loader's guard (demo.DecideLoad) then refused the database as a real one, and the demo stayed stuck without View demo. The author approved this demo-only change in chat on 2026-11-03, relayed by the orchestrator; it is not in DESIGN.md yet, so this requirement has default firmness. It narrows, for demo deployments only, the firm REQ-ui-first-run-wizard and REQ-identity-first-run-wizard-follows and the default REQ-ui-first-run-wizard-steps: an install without demo sessions keeps the wizard unchanged.

## Source

> Author-approved demo-only change: a demo deployment never shows the first-run setup wizard; until the demo marker exists, the front door shows a "demo data loading" state instead.

Author decision given in chat (2026-11-03); docs/dev/demo.md#demo-data
