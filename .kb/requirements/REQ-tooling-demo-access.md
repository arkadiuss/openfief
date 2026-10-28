---
title: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when no key has that label.
status: open
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
semantic_text: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when no key has that label.
logic_claims:
  - CLAIM-895A6F8BE6B71F16
semantic_clauses:
  - A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when no key has that label.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bb0e195ac5a2f3a27e656771d130bcab645e318150a90dc036065a67eeb2c86b
semantic_inventory:
  - claim_key: CLAIM-895A6F8BE6B71F16
    claim_text: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when no key has that label
    role: normative
    status: modeled
    span:
      start: 0
      end: 316
origin:
  kind: agent
  recorded_at: '2026-10-28T09:00:42.816Z'
id: REQ-tooling-demo-access
type: req
---
A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when no key has that label.

## Context

SSH carries administration and port-forward tunnels, so it stays limited to the owner's addresses and keys; there is no password to leak. HTTP and HTTPS reach k3s's Traefik for the demo's routes and Let's Encrypt's HTTP-01 challenge, open by default because a demo holds synthetic data only (GR-12). Modeled by Claude Code for the demo machine automation, which the orchestrating agent specified from the author's request for good committed automation of the demo machine, easy to create and recreate, with spending limited to 10 EUR a month if possible; the author has not reviewed it yet. The author later asked that the demo take normal HTTP traffic through k3s's own Traefik, with cert-manager and Let's Encrypt on an sslip.io name, while the product bases keep shipping Services only (D-28 unchanged for them).

## Source

> Demo and test environment: a Hetzner machine running k3s, deployed by the pipeline from `main` and from beta releases. It holds synthetic data only: never real broker credentials or portfolio data (GR-12).

docs/dev/sdlc.md#ci-and-environments
