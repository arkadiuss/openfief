---
title: A core running with demo sessions on must report, while the database does not carry the demo marker yet, that the demo data is loading, and must refuse every change of sleeves, sleeve kinds and accounts until the marker exists, so that nothing the first-run wizard would write can be stored on a demo before its dataset is loaded.
status: open
text_ref: docs/dev/demo.md#demo-data
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-demo-data
  - lane:strict
semantic_text: A core running with demo sessions on must report, while the database does not carry the demo marker yet, that the demo data is loading, and must refuse every change of sleeves, sleeve kinds and accounts until the marker exists, so that nothing the first-run wizard would write can be stored on a demo before its dataset is loaded.
logic_claims:
  - CLAIM-A3CB7C78D051943B
semantic_clauses:
  - A core running with demo sessions on must report, while the database does not carry the demo marker yet, that the demo data is loading, and must refuse every change of sleeves, sleeve kinds and accounts until the marker exists, so that nothing the first-run wizard would write can be stored on a demo before its dataset is loaded.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ec263623500c2c5bbac5fc4a11c923bde9c3b3595a5f430900dcb6a9128f1290
semantic_inventory:
  - claim_key: CLAIM-A3CB7C78D051943B
    claim_text: A core running with demo sessions on must report, while the database does not carry the demo marker yet, that the demo data is loading, and must refuse every change of sleeves, sleeve kinds and accounts until the marker exists, so that nothing the first-run wizard would write can be stored on a demo before its dataset is loaded
    role: normative
    status: modeled
    span:
      start: 0
      end: 329
origin:
  kind: agent
  recorded_at: '2026-11-03T19:04:14.058Z'
id: REQ-identity-demo-loading-no-wizard
type: req
---
A core running with demo sessions on must report, while the database does not carry the demo marker yet, that the demo data is loading, and must refuse every change of sleeves, sleeve kinds and accounts until the marker exists, so that nothing the first-run wizard would write can be stored on a demo before its dataset is loaded.

## Context

On a fresh demo database, before cmd/demo-seed has written the demo marker, a visitor who enrolled a passkey landed in the first-run wizard, whose sleeves and accounts are data without the marker; the loader's guard (demo.DecideLoad) then refused the database as a real one, and the demo stayed stuck without View demo. The author approved this demo-only change in chat on 2026-11-03, relayed by the orchestrator; it is not in DESIGN.md yet, so this requirement has default firmness. It narrows, for demo deployments only, the firm REQ-ui-first-run-wizard and REQ-identity-first-run-wizard-follows and the default REQ-ui-first-run-wizard-steps: an install without demo sessions keeps the wizard unchanged.

## Source

> Author-approved demo-only change: a demo deployment never shows the first-run setup wizard; until the demo marker exists, the front door shows a "demo data loading" state instead.

Author decision given in chat (2026-11-03); docs/dev/demo.md#demo-data
