---
title: 'First-run setup must happen in the built-in UI: after the owner creates the first passkey on an install that is not a demo, a first-run wizard must walk them through creating accounts and sleeves and binding the connector instances already installed.'
status: open
text_ref: design-decisions:D-191#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-191-1
  - lane:strict
semantic_text: 'First-run setup must happen in the built-in UI: after the owner creates the first passkey on an install that is not a demo, a first-run wizard must walk them through creating accounts and sleeves and binding the connector instances already installed.'
logic_claims:
  - CLAIM-F20CA6402781A17D
semantic_clauses:
  - 'First-run setup must happen in the built-in UI: after the owner creates the first passkey on an install that is not a demo, a first-run wizard must walk them through creating accounts and sleeves and binding the connector instances already installed.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f40aa2edec6f0ec3650e035fae76e272edd00f6a07d07372442fe4f53c762da1
semantic_inventory:
  - claim_key: CLAIM-F20CA6402781A17D
    claim_text: 'First-run setup must happen in the built-in UI: after the owner creates the first passkey on an install that is not a demo, a first-run wizard must walk them through creating accounts and sleeves and binding the connector instances already installed'
    role: normative
    status: modeled
    span:
      start: 0
      end: 249
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-191#1
  recorded_at: '2026-11-05T23:58:57.534Z'
id: REQ-ui-first-run-wizard-outside-demo
type: req
---
First-run setup must happen in the built-in UI: after the owner creates the first passkey on an install that is not a demo, a first-run wizard must walk them through creating accounts and sleeves and binding the connector instances already installed.

## Context

D-73 put first-run setup in the built-in UI after the first passkey (D-60): accounts (D-72), sleeves and binding the connector instances already installed, while installing plugins stays in the CLI (D-46), whose last step points to the wizard (REQ-plugins-cli-install-points-to-wizard). D-191 leaves that unchanged for every install but a demo.

## Why this changed

D-191 (author decision, 2026-11-05): a demo shows no first-run wizard. On a fresh demo database, before cmd/demo-seed wrote the demo marker, a visitor who enrolled a passkey landed in the wizard; the sleeves and accounts it stored were data without the marker, the loader's guard then refused the database as a real one, and the demo stayed stuck without View demo. So the wizard now belongs to installs that are not a demo; on a demo the front door says the demo data is loading until the marker exists (REQ-demo-no-wizard-until-loaded).

## Source

> **[D-73]** First-run setup happens in the built-in UI. After the owner creates the first passkey (D-60), a first-run wizard walks them through creating accounts (D-72), sleeves, and binding the connector instances already installed. Installing plugins stays in the CLI (D-46), whose last step points to the wizard.

> **[D-191]** A demo shows no first-run wizard: until the demo dataset's marker exists, it shows that the demo data is loading, so nothing entered before the seed runs can stop it from loading. Once loaded, the owner can still sign in and change it (D-68) until the next reseed.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-191#1
