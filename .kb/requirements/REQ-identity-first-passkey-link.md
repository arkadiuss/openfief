---
title: The openfief CLI's first-passkey command must start the kubectl port-forward to the UI and print D-60's one-time link, and the first-run wizard must follow.
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
  recorded_at: '2026-10-29T08:10:34.672Z'
semantic_text: The openfief CLI's first-passkey command must start the kubectl port-forward to the UI and print D-60's one-time link, and the first-run wizard must follow.
logic_claims:
  - CLAIM-AF850D4D91595E74
semantic_clauses:
  - The openfief CLI's first-passkey command must start the kubectl port-forward to the UI and print D-60's one-time link, and the first-run wizard must follow.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 059561d3b87f2099d02a9d39264802942b379ac4929cdd6551ab4d5b2af160c7
semantic_inventory:
  - claim_key: CLAIM-AF850D4D91595E74
    claim_text: The openfief CLI's first-passkey command must start the kubectl port-forward to the UI and print D-60's one-time link, and the first-run wizard must follow
    role: normative
    status: modeled
    span:
      start: 0
      end: 155
id: REQ-identity-first-passkey-link
type: req
---
The openfief CLI's first-passkey command must start the kubectl port-forward to the UI and print D-60's one-time link, and the first-run wizard must follow.

## Context

D-60 creates the owner's first passkey on localhost through kubectl port-forward, with a one-time link the CLI prints (REQ-identity-install-without-hostname); D-80's command runs the core's enrollment command for it (REQ-identity-first-passkey-command). After the first passkey, D-73's first-run wizard in the built-in UI walks the owner through setup (REQ-ui-first-run-wizard). D-81, recorded after D-80, makes the core the only entry point for browsers, so kubectl port-forward needs to reach the core alone (REQ-core-browser-entry-point). The author recorded D-79 and D-80 after the design session that answered the implementation's questions.

## Source

> The `openfief` CLI has a command for the first passkey: it runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig, starts the `kubectl port-forward` to the UI, and prints D-60's one-time link. The first-run wizard (D-73) follows.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-80#2
