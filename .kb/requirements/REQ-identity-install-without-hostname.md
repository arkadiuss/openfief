---
title: OpenFief must install with no hostname, and the owner's first passkey must be created on localhost through a one-time link the CLI prints.
status: open
text_ref: design-decisions:D-60#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-60-1
  - lane:strict
semantic_text: OpenFief must install with no hostname, and the owner's first passkey must be created on localhost through a one-time link the CLI prints.
logic_claims:
  - CLAIM-7DD67F5DF8CCE695
semantic_clauses:
  - OpenFief must install with no hostname, and the owner's first passkey must be created on localhost through a one-time link the CLI prints.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 99ddd193af07a288c68e0011cdd305a4aea35e8f4832ed805d65616c6c8aaef9
semantic_inventory:
  - claim_key: CLAIM-7DD67F5DF8CCE695
    claim_text: OpenFief must install with no hostname, and the owner's first passkey must be created on localhost through a one-time link the CLI prints
    role: normative
    status: modeled
    span:
      start: 0
      end: 137
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-60#1
  recorded_at: '2026-10-28T06:09:44.518Z'
id: REQ-identity-install-without-hostname
type: req
---
OpenFief must install with no hostname, and the owner's first passkey must be created on localhost through a one-time link the CLI prints.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files.

## Source

> OpenFief installs with no hostname. A passkey works only on the hostname it was created on, so the CLI prints a one-time link with which the owner creates the first passkey on `localhost`, through `kubectl port-forward`. When the owner adds their own Ingress, a logged-in session issues a one-time link for creating a passkey on the new hostname. Passkeys are kept per hostname.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-60#1
