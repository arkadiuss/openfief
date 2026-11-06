---
title: The built-in UI must never show the first-run wizard on a demo deployment, redirecting its pages to the overview, and while the demo data is loading its login page must say so and ask the visitor to refresh in a minute instead of offering View demo.
status: open
text_ref: docs/DESIGN.md#D-191
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

On a fresh demo database, before cmd/demo-seed has written the demo marker, a visitor who enrolled a passkey landed in the first-run wizard, whose sleeves and accounts are data without the marker; the loader's guard (demo.DecideLoad) then refused the database as a real one, and the demo stayed stuck without View demo. The author approved this demo-only change in chat on 2026-11-03 and recorded it as D-191. This requirement is the implementation's reading of D-191, with default firmness; the firm REQ-demo-no-wizard-until-loaded and REQ-demo-owner-changes-after-load import D-191 itself. With it, REQ-ui-first-run-wizard-outside-demo, REQ-identity-first-run-wizard-follows-outside-demo and REQ-ui-first-run-wizard-steps-outside-demo superseded the wizard requirements: an install without demo sessions keeps the wizard unchanged.

## Source

> **[D-191]** A demo shows no first-run wizard: until the demo dataset's marker exists, it shows that the demo data is loading, so nothing entered before the seed runs can stop it from loading. Once loaded, the owner can still sign in and change it (D-68) until the next reseed.

docs/DESIGN.md#D-191; docs/dev/demo.md#demo-data
