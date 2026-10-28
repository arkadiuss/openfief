---
title: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when SSH sources are allowed but no key has that label, or when DEMO_SSH=keys finds no such key.
status: open
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
semantic_text: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when SSH sources are allowed but no key has that label, or when DEMO_SSH=keys finds no such key.
logic_claims:
  - CLAIM-15F096C98BCFD6AB
semantic_clauses:
  - A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when SSH sources are allowed but no key has that label, or when DEMO_SSH=keys finds no such key.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 73cfb006871ff58e8232cf945b2ed0161a24752f4d240fd3a83dd719e892b7ae
semantic_inventory:
  - claim_key: CLAIM-15F096C98BCFD6AB
    claim_text: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when SSH sources are allowed but no key has that label, or when DEMO_SSH=keys finds no such key
    role: normative
    status: modeled
    span:
      start: 0
      end: 385
origin:
  kind: agent
  recorded_at: '2026-10-28T09:00:42.816Z'
id: REQ-tooling-demo-access
type: req
---
A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when SSH sources are allowed but no key has that label, or when DEMO_SSH=keys finds no such key.

## Context

SSH carries administration and port-forward tunnels, so it stays limited to the owner's addresses and keys; there is no password to leak. HTTP and HTTPS reach k3s's Traefik for the demo's routes and Let's Encrypt's HTTP-01 challenge, open by default because a demo holds synthetic data only (GR-12). Without any key and with no SSH source, REQ-tooling-demo-ssh-throwaway applies instead of a refusal. Modeled by Claude Code for the demo machine automation. The author asked, through the orchestrating agent, for the demo on the real domain openfief.dev (a Cloudflare zone on the Free plan), fully automated, and for an explicit SSH-off mode because the Hetzner project has no SSH key and Hetzner would otherwise email a root password. The docs site shares the zone (docs.openfief.dev), so the demo must touch only its own records. The author has not reviewed this requirement yet.

## Source

> Demo and test environment: a Hetzner machine running k3s, deployed by the pipeline from `main` and from beta releases. It holds synthetic data only: never real broker credentials or portfolio data (GR-12).

docs/dev/sdlc.md#ci-and-environments
