---
title: Without an origin, the openfief CLI's first-passkey command must forward a local port to the core and print D-60's one-time link for localhost.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-80-3
  - lane:strict
text_ref: design-decisions:D-80#3
origin:
  kind: import
  ref: design-decisions:D-80#3
  recorded_at: '2026-10-29T15:05:03.288Z'
semantic_text: Without an origin, the openfief CLI's first-passkey command must forward a local port to the core and print D-60's one-time link for localhost.
logic_claims:
  - CLAIM-06081FE0F074F623
semantic_clauses:
  - Without an origin, the openfief CLI's first-passkey command must forward a local port to the core and print D-60's one-time link for localhost.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2447875e3a05a4c25630188bc9ebbc4241ebe748662ee843f089a51775201182
semantic_inventory:
  - claim_key: CLAIM-06081FE0F074F623
    claim_text: Without an origin, the openfief CLI's first-passkey command must forward a local port to the core and print D-60's one-time link for localhost
    role: normative
    status: modeled
    span:
      start: 0
      end: 142
id: REQ-identity-enroll-localhost-forward
type: req
---
Without an origin, the openfief CLI's first-passkey command must forward a local port to the core and print D-60's one-time link for localhost.

## Why this changed

REQ-identity-first-passkey-link had the first-passkey command always start a `kubectl port-forward` to the UI and print the link for `localhost`. In the docs session of 2026-10-29 the author changed D-28 and D-80: `kubectl port-forward` is no longer the default way in, OpenFief ships the core's ClusterIP Service plus an optional Gateway API HTTPRoute component (REQ-platform-httproute-component), and enroll can target the owner's hostname when given an origin (REQ-identity-enroll-origin). The local port forward is kept only for the case without an origin, and it goes to the core, not the UI, since the core is the only entry point for browsers (REQ-core-only-browser-entry-point).

## Context

D-60 keeps `localhost` through a port-forward for an owner who has no route yet (REQ-identity-first-passkey-own-host-or-localhost). The command runs the core's enrollment command through Kubernetes exec (REQ-identity-first-passkey-command). The author decided this when reviewing the docs, which over-emphasized `kubectl port-forward`.

## Source

> The `openfief` CLI has a command for the first passkey: it runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig, and prints D-60's one-time link. Given an origin, such as `https://fief.example.org`, the link is for that origin, which must be one the core allows; without one, the command forwards a local port to the core and the link is for `localhost`. The first-run wizard (D-73) follows.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-80#3
