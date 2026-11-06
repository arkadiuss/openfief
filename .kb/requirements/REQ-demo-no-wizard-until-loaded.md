---
title: A demo must show no first-run wizard, and until the demo dataset's marker exists it must show that the demo data is loading, so that nothing entered before the seed runs can stop the demo data from loading.
status: open
text_ref: design-decisions:D-191#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-191-1
  - lane:strict
semantic_text: A demo must show no first-run wizard, and until the demo dataset's marker exists it must show that the demo data is loading, so that nothing entered before the seed runs can stop the demo data from loading.
logic_claims:
  - CLAIM-C5454EEBE9B4A698
semantic_clauses:
  - A demo must show no first-run wizard, and until the demo dataset's marker exists it must show that the demo data is loading, so that nothing entered before the seed runs can stop the demo data from loading.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9a5c5355ce468b16b005e894508402845e5547f3dc55d969ceefd03f9e2ed3d8
semantic_inventory:
  - claim_key: CLAIM-C5454EEBE9B4A698
    claim_text: A demo must show no first-run wizard, and until the demo dataset's marker exists it must show that the demo data is loading, so that nothing entered before the seed runs can stop the demo data from loading
    role: normative
    status: modeled
    span:
      start: 0
      end: 205
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-191#1
  recorded_at: '2026-11-05T23:58:44.932Z'
id: REQ-demo-no-wizard-until-loaded
type: req
---
A demo must show no first-run wizard, and until the demo dataset's marker exists it must show that the demo data is loading, so that nothing entered before the seed runs can stop the demo data from loading.

## Context

A demo is a core running with demo sessions on (--demo-sessions, D-68), whose data cmd/demo-seed loads and marks. Anything the owner enters before the marker exists, such as a wizard's sleeves and accounts, would make the loader refuse the database as a real one. The default refinements REQ-ui-demo-no-wizard (the built-in UI's pages) and REQ-identity-demo-loading-no-wizard (the core's report and its refusal of sleeve and account changes) say how. The author recorded D-191 after the demo got stuck that way; no further reason is stated.

## Source

> **[D-191]** A demo shows no first-run wizard: until the demo dataset's marker exists, it shows that the demo data is loading, so nothing entered before the seed runs can stop it from loading. Once loaded, the owner can still sign in and change it (D-68) until the next reseed.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-191#1
