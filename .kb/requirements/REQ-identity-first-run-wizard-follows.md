---
title: The first-run wizard must follow the creation of the first passkey.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-80-4
  - lane:strict
text_ref: design-decisions:D-80#4
origin:
  kind: import
  ref: design-decisions:D-80#4
  recorded_at: '2026-10-29T15:05:03.288Z'
semantic_text: The first-run wizard must follow the creation of the first passkey.
logic_claims:
  - CLAIM-D7AF92FFE14F8E01
semantic_clauses:
  - The first-run wizard must follow the creation of the first passkey.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 636dcfdf3ad23c8f57ddba83819d4e32da16c3d6b40e7ff6b91cd190b9800cb2
semantic_inventory:
  - claim_key: CLAIM-D7AF92FFE14F8E01
    claim_text: The first-run wizard must follow the creation of the first passkey
    role: normative
    status: modeled
    span:
      start: 0
      end: 66
id: REQ-identity-first-run-wizard-follows
type: req
---
The first-run wizard must follow the creation of the first passkey.

## Why this changed

REQ-identity-first-passkey-link bundled this with starting a `kubectl port-forward` to the UI and printing the link for `localhost`. In the docs session of 2026-10-29 the author changed D-28 and D-80: `kubectl port-forward` is no longer the default way in, OpenFief ships the core's ClusterIP Service plus an optional Gateway API HTTPRoute component (REQ-platform-httproute-component), and enroll can target the owner's hostname when given an origin. The wizard still follows the first passkey on either path, so that part of the old requirement stands here on its own, next to REQ-identity-enroll-origin and REQ-identity-enroll-localhost-forward.

## Context

After the first passkey, D-73's first-run wizard in the built-in UI walks the owner through creating accounts and sleeves and binding the connector instances already installed (REQ-ui-first-run-wizard), and plugin installation in the CLI ends by pointing to it (REQ-plugins-cli-install-points-to-wizard). The author decided this when reviewing the docs, which over-emphasized `kubectl port-forward`.

## Source

> The `openfief` CLI has a command for the first passkey: it runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig, and prints D-60's one-time link. Given an origin, such as `https://fief.example.org`, the link is for that origin, which must be one the core allows; without one, the command forwards a local port to the core and the link is for `localhost`. The first-run wizard (D-73) follows.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-80#4
