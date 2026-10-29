---
title: OpenFief must install with no hostname, and the owner's first passkey must be created through a one-time link the CLI prints, on the owner's own hostname or on localhost.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-60-1
  - lane:strict
text_ref: design-decisions:D-60#1
origin:
  kind: import
  ref: design-decisions:D-60#1
  recorded_at: '2026-10-29T15:05:03.288Z'
semantic_text: OpenFief must install with no hostname, and the owner's first passkey must be created through a one-time link the CLI prints, on the owner's own hostname or on localhost.
logic_claims:
  - CLAIM-5D12A8F1F5139BBF
semantic_clauses:
  - OpenFief must install with no hostname, and the owner's first passkey must be created through a one-time link the CLI prints, on the owner's own hostname or on localhost.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 17bb52821e48ca702c7f7a1cda1c88fa364dd9d9d94d1e4f56b3fdf301c1838c
semantic_inventory:
  - claim_key: CLAIM-5D12A8F1F5139BBF
    claim_text: OpenFief must install with no hostname, and the owner's first passkey must be created through a one-time link the CLI prints, on the owner's own hostname or on localhost
    role: normative
    status: modeled
    span:
      start: 0
      end: 169
id: REQ-identity-first-passkey-own-host-or-localhost
type: req
---
OpenFief must install with no hostname, and the owner's first passkey must be created through a one-time link the CLI prints, on the owner's own hostname or on localhost.

## Why this changed

REQ-identity-install-without-hostname created the first passkey on `localhost` only, through `kubectl port-forward`, which D-28 then made the default way in. In the docs session of 2026-10-29 the author changed D-28 and D-60: `kubectl port-forward` is no longer the default way in, OpenFief ships the core's ClusterIP Service plus an optional Gateway API HTTPRoute component (REQ-platform-httproute-component), and the first-passkey link can be for the owner's own hostname, with `localhost` through a port-forward kept for an owner who has no route yet.

## Context

A passkey works only on the hostname it was created on, and OpenFief installs with no hostname, so the CLI prints the link for the hostname the owner will use. D-80's command prints it for an origin the core allows (REQ-identity-enroll-origin) or forwards a local port and prints it for `localhost` (REQ-identity-enroll-localhost-forward). A logged-in session issues links for further hostnames (REQ-identity-passkey-new-hostname). The author decided this when reviewing the docs, which over-emphasized `kubectl port-forward`.

## Source

> OpenFief installs with no hostname. A passkey works only on the hostname it was created on, so the CLI prints a one-time link for the first passkey: on the owner's own hostname, or on `localhost` through a port-forward while the owner has no route. A logged-in session issues a one-time link for creating a passkey on another hostname. Passkeys are kept per hostname.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-60#1
