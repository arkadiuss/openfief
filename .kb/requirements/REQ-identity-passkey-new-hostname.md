---
title: A logged-in session must be able to issue a one-time link for creating a passkey on a new hostname, and passkeys must be kept per hostname.
status: open
text_ref: design-decisions:D-60#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-60-2
  - lane:strict
semantic_text: A logged-in session must be able to issue a one-time link for creating a passkey on a new hostname, and passkeys must be kept per hostname.
logic_claims:
  - CLAIM-207A16FB308FD2F2
semantic_clauses:
  - A logged-in session must be able to issue a one-time link for creating a passkey on a new hostname, and passkeys must be kept per hostname.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0b021fe24d31725881c4f926d0dcbe7942ae62a2a2ef7626188ff8facfa7636e
semantic_inventory:
  - claim_key: CLAIM-207A16FB308FD2F2
    claim_text: A logged-in session must be able to issue a one-time link for creating a passkey on a new hostname, and passkeys must be kept per hostname
    role: normative
    status: modeled
    span:
      start: 0
      end: 138
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-60#2
  recorded_at: '2026-10-28T06:09:47.643Z'
id: REQ-identity-passkey-new-hostname
type: req
---
A logged-in session must be able to issue a one-time link for creating a passkey on a new hostname, and passkeys must be kept per hostname.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files.

## Source

> OpenFief installs with no hostname. A passkey works only on the hostname it was created on, so the CLI prints a one-time link with which the owner creates the first passkey on `localhost`, through `kubectl port-forward`. When the owner adds their own Ingress, a logged-in session issues a one-time link for creating a passkey on the new hostname. Passkeys are kept per hostname.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-60#2
