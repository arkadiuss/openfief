---
title: A core running with demo sessions on must report, while the database does not carry the demo marker yet, that the demo data is loading, and must refuse every change of sleeves, sleeve kinds and accounts until the marker exists, so that nothing the first-run wizard would write can be stored on a demo before its dataset is loaded.
status: open
text_ref: docs/DESIGN.md#D-191
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

On a fresh demo database, before cmd/demo-seed has written the demo marker, a visitor who enrolled a passkey landed in the first-run wizard, whose sleeves and accounts are data without the marker; the loader's guard (demo.DecideLoad) then refused the database as a real one, and the demo stayed stuck without View demo. The author approved this demo-only change in chat on 2026-11-03 and recorded it as D-191. This requirement is the implementation's reading of D-191, with default firmness; the firm REQ-demo-no-wizard-until-loaded and REQ-demo-owner-changes-after-load import D-191 itself. With it, REQ-ui-first-run-wizard-outside-demo, REQ-identity-first-run-wizard-follows-outside-demo and REQ-ui-first-run-wizard-steps-outside-demo superseded the wizard requirements: an install without demo sessions keeps the wizard unchanged.

## Source

> **[D-191]** A demo shows no first-run wizard: until the demo dataset's marker exists, it shows that the demo data is loading, so nothing entered before the seed runs can stop it from loading. Once loaded, the owner can still sign in and change it (D-68) until the next reseed.

docs/DESIGN.md#D-191; docs/dev/demo.md#demo-data
