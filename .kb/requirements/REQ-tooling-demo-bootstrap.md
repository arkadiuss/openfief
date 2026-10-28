---
title: A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, write the demo settings (the demo's host, <ENV>.<DEMO_DOMAIN> or with DEMO_DOMAIN=none the machine's sslip.io name, and the Let's Encrypt production or staging server) and no secrets, and rendering must fail when a pinned value or setting is missing or malformed or a placeholder is left.
status: open
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
semantic_text: A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, write the demo settings (the demo's host, <ENV>.<DEMO_DOMAIN> or with DEMO_DOMAIN=none the machine's sslip.io name, and the Let's Encrypt production or staging server) and no secrets, and rendering must fail when a pinned value or setting is missing or malformed or a placeholder is left.
logic_claims:
  - CLAIM-0DD96500B0521177
semantic_clauses:
  - A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, write the demo settings (the demo's host, <ENV>.<DEMO_DOMAIN> or with DEMO_DOMAIN=none the machine's sslip.io name, and the Let's Encrypt production or staging server) and no secrets, and rendering must fail when a pinned value or setting is missing or malformed or a placeholder is left.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5fa45c294b438840db2d8f487ce76df02842ea7e470339cfa46b06b5d3f4560b
semantic_inventory:
  - claim_key: CLAIM-0DD96500B0521177
    claim_text: A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, write the demo settings (the demo's host, <ENV>.<DEMO_DOMAIN> or with DEMO_DOMAIN=none the machine's sslip.io name, and the Let's Encrypt production or staging server) and no secrets, and rendering must fail when a pinned value or setting is missing or malformed or a placeholder is left
    role: normative
    status: modeled
    span:
      start: 0
      end: 427
origin:
  kind: agent
  recorded_at: '2026-10-28T09:01:10.232Z'
id: REQ-tooling-demo-bootstrap
type: req
---
A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, write the demo settings (the demo's host, <ENV>.<DEMO_DOMAIN> or with DEMO_DOMAIN=none the machine's sslip.io name, and the Let's Encrypt production or staging server) and no secrets, and rendering must fail when a pinned value or setting is missing or malformed or a placeholder is left.

## Context

Pinning and checksums make a recreated machine identical and stop a tampered download. D-12 needs a CNI that enforces NetworkPolicy, which k3s's flannel with its policy controller provides. cloud-init user data is readable from the machine's metadata service, so it must hold no secret; with DEMO_DOMAIN=none the host name comes from the machine's own IPv4, known only once it exists. Modeled by Claude Code for the demo machine automation, which the orchestrating agent specified from the author's request for good committed automation of the demo machine, easy to create and recreate, with spending limited to 10 EUR a month if possible; the author has not reviewed it yet. The author later asked that the demo take normal HTTP traffic through k3s's own Traefik, with cert-manager and Let's Encrypt on an sslip.io name, while the product bases keep shipping Services only (D-28 unchanged for them). The author then bought the domain openfief.dev for the demo, the docs and a landing page, so the demo is served at <ENV>.<DEMO_DOMAIN>, demo.openfief.dev by default, and sslip.io stays as the fallback that needs no DNS record.

## Source

> Tested on kind and k3d, with documented requirements: a CNI that enforces NetworkPolicy, and a storage class.

docs/DESIGN.md D-12

> Demo and test environment: a Hetzner machine running k3s, deployed by the pipeline from `main` and from beta releases. It holds synthetic data only: never real broker credentials or portfolio data (GR-12).

docs/dev/sdlc.md#ci-and-environments

> Bought you a domain openfief.dev to make demo docs and landing

The author, in the design session, 2026-10-28
