---
title: On an install that is not a demo, the first-run wizard must follow the creation of the first passkey.
status: open
text_ref: design-decisions:D-191#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-191-1
  - lane:strict
semantic_text: On an install that is not a demo, the first-run wizard must follow the creation of the first passkey.
logic_claims:
  - CLAIM-4D6D829E84C65AF6
semantic_clauses:
  - On an install that is not a demo, the first-run wizard must follow the creation of the first passkey.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 967396da0aa913df46c1961bd2544ed32dd5962880a88d0025b396dac2cf5c61
semantic_inventory:
  - claim_key: CLAIM-4D6D829E84C65AF6
    claim_text: On an install that is not a demo, the first-run wizard must follow the creation of the first passkey
    role: normative
    status: modeled
    span:
      start: 0
      end: 100
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-191#1
  recorded_at: '2026-11-05T23:59:05.191Z'
id: REQ-identity-first-run-wizard-follows-outside-demo
type: req
---
On an install that is not a demo, the first-run wizard must follow the creation of the first passkey.

## Context

D-80 has the openfief CLI print the first passkey's one-time link, for the owner's origin or for localhost (REQ-identity-enroll-origin, REQ-identity-enroll-localhost-forward), and the first-run wizard (D-73) follows on either path. D-191 keeps that for every install but a demo, where the owner who enrolls a passkey lands on the overview instead.

## Why this changed

D-191 (author decision, 2026-11-05): a demo shows no first-run wizard. On a fresh demo database, before cmd/demo-seed wrote the demo marker, a visitor who enrolled a passkey landed in the wizard; the sleeves and accounts it stored were data without the marker, the loader's guard then refused the database as a real one, and the demo stayed stuck without View demo. So the wizard now belongs to installs that are not a demo; on a demo the front door says the demo data is loading until the marker exists (REQ-demo-no-wizard-until-loaded).

## Source

> **[D-80]** ... Given an origin, such as `https://fief.example.org`, the link is for that origin, which must be one the core allows; without one, the command forwards a local port to the core and the link is for `localhost`. The first-run wizard (D-73) follows.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-80#4

> **[D-191]** A demo shows no first-run wizard: until the demo dataset's marker exists, it shows that the demo data is loading, so nothing entered before the seed runs can stop it from loading. Once loaded, the owner can still sign in and change it (D-68) until the next reseed.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-191#1
