---
title: Given an origin the core allows, the openfief CLI's first-passkey command must print D-60's one-time link for that origin without forwarding a port.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-80-2
  - lane:strict
text_ref: design-decisions:D-80#2
origin:
  kind: import
  ref: design-decisions:D-80#2
  recorded_at: '2026-10-29T15:05:03.288Z'
semantic_text: Given an origin the core allows, the openfief CLI's first-passkey command must print D-60's one-time link for that origin without forwarding a port.
logic_claims:
  - CLAIM-C8E79E749AD6D89E
semantic_clauses:
  - Given an origin the core allows, the openfief CLI's first-passkey command must print D-60's one-time link for that origin without forwarding a port.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 0ad6420a3bf9c725926bccda662593ca7e3bb1969a906904995ef07a539e5f97
semantic_inventory:
  - claim_key: CLAIM-C8E79E749AD6D89E
    claim_text: Given an origin the core allows, the openfief CLI's first-passkey command must print D-60's one-time link for that origin without forwarding a port
    role: normative
    status: modeled
    span:
      start: 0
      end: 147
id: REQ-identity-enroll-origin
type: req
---
Given an origin the core allows, the openfief CLI's first-passkey command must print D-60's one-time link for that origin without forwarding a port.

## Why this changed

REQ-identity-first-passkey-link had the first-passkey command always start a `kubectl port-forward` to the UI and print the link for `localhost`, with the first-run wizard following. In the docs session of 2026-10-29 the author changed D-28 and D-80: `kubectl port-forward` is no longer the default way in, OpenFief ships the core's ClusterIP Service plus an optional Gateway API HTTPRoute component (REQ-platform-httproute-component), and enroll can target the owner's hostname when given an origin. The old requirement is split in three: this one for a given origin, REQ-identity-enroll-localhost-forward without one, and REQ-identity-first-run-wizard-follows.

## Context

D-60 lets the owner create the first passkey on their own hostname (REQ-identity-first-passkey-own-host-or-localhost), and a passkey works only on the hostname it was created on. The core accepts the owner's browser only on origins on its allow-list (REQ-identity-origins-per-host), so the given origin must be one of those. The command runs the core's enrollment command through Kubernetes exec (REQ-identity-first-passkey-command). The author decided this when reviewing the docs, which over-emphasized `kubectl port-forward`.

## Source

> The `openfief` CLI has a command for the first passkey: it runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig, and prints D-60's one-time link. Given an origin, such as `https://fief.example.org`, the link is for that origin, which must be one the core allows; without one, the command forwards a local port to the core and the link is for `localhost`. The first-run wizard (D-73) follows.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-80#2
