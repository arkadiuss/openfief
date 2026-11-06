---
title: Once a demo's data is loaded, the owner must still be able to sign in and change it until the next reseed.
status: open
text_ref: design-decisions:D-191#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-191-2
  - lane:strict
semantic_text: Once a demo's data is loaded, the owner must still be able to sign in and change it until the next reseed.
logic_claims:
  - CLAIM-F1F1E6499B8629A3
semantic_clauses:
  - Once a demo's data is loaded, the owner must still be able to sign in and change it until the next reseed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b90a1d7d12465955e3123cb048e8deaa4be0d0e5e3d749f91742e725807d8510
semantic_inventory:
  - claim_key: CLAIM-F1F1E6499B8629A3
    claim_text: Once a demo's data is loaded, the owner must still be able to sign in and change it until the next reseed
    role: normative
    status: modeled
    span:
      start: 0
      end: 105
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-191#2
  recorded_at: '2026-11-05T23:58:50.457Z'
id: REQ-demo-owner-changes-after-load
type: req
---
Once a demo's data is loaded, the owner must still be able to sign in and change it until the next reseed.

## Context

D-68 lets the owner sign in to the demo with a passkey next to the read-only demo sessions. D-191 keeps that: the refusal of sleeve and account changes lasts only while the demo data is loading (REQ-identity-demo-loading-no-wizard), and the next reseed replaces whatever the owner changed. The author recorded it with the first sentence of D-191; no further reason is stated.

## Source

> **[D-191]** A demo shows no first-run wizard: until the demo dataset's marker exists, it shows that the demo data is loading, so nothing entered before the seed runs can stop it from loading. Once loaded, the owner can still sign in and change it (D-68) until the next reseed.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-191#2
