---
title: With CLOUDFLARE_API_TOKEN set and DEMO_DOMAIN not none, the demo automation must point the demo's host name <ENV>.<DEMO_DOMAIN> (demo.openfief.dev for the demo named demo) at its server with a DNS-only A record and a CAA record that allows only letsencrypt.org, each with the comment managed-by=openfief-demo env=<env>, creating missing records and updating changed ones in place, keeping them across recreate, and deleting them on down; it must never update or delete a record without that comment, and must refuse when such a record occupies the host name.
status: open
text_ref: docs/dev/demo.md#dns
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-dns
  - lane:strict
semantic_text: With CLOUDFLARE_API_TOKEN set and DEMO_DOMAIN not none, the demo automation must point the demo's host name <ENV>.<DEMO_DOMAIN> (demo.openfief.dev for the demo named demo) at its server with a DNS-only A record and a CAA record that allows only letsencrypt.org, each with the comment managed-by=openfief-demo env=<env>, creating missing records and updating changed ones in place, keeping them across recreate, and deleting them on down; it must never update or delete a record without that comment, and must refuse when such a record occupies the host name.
logic_claims:
  - CLAIM-8AE21BA437332713
semantic_clauses:
  - With CLOUDFLARE_API_TOKEN set and DEMO_DOMAIN not none, the demo automation must point the demo's host name <ENV>.<DEMO_DOMAIN> (demo.openfief.dev for the demo named demo) at its server with a DNS-only A record and a CAA record that allows only letsencrypt.org, each with the comment managed-by=openfief-demo env=<env>, creating missing records and updating changed ones in place, keeping them across recreate, and deleting them on down; it must never update or delete a record without that comment, and must refuse when such a record occupies the host name.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9769f78533ddb0ce0640afb199fe53972ffdabdd72bf4828b952d68460dde1a8
semantic_inventory:
  - claim_key: CLAIM-8AE21BA437332713
    claim_text: With CLOUDFLARE_API_TOKEN set and DEMO_DOMAIN not none, the demo automation must point the demo's host name <ENV>.<DEMO_DOMAIN> (demo.openfief.dev for the demo named demo) at its server with a DNS-only A record and a CAA record that allows only letsencrypt.org, each with the comment managed-by=openfief-demo env=<env>, creating missing records and updating changed ones in place, keeping them across recreate, and deleting them on down; it must never update or delete a record without that comment, and must refuse when such a record occupies the host name
    role: normative
    status: modeled
    span:
      start: 0
      end: 557
origin:
  kind: agent
  recorded_at: '2026-10-28T14:35:46.704Z'
id: REQ-tooling-demo-dns
type: req
---
With CLOUDFLARE_API_TOKEN set and DEMO_DOMAIN not none, the demo automation must point the demo's host name <ENV>.<DEMO_DOMAIN> (demo.openfief.dev for the demo named demo) at its server with a DNS-only A record and a CAA record that allows only letsencrypt.org, each with the comment managed-by=openfief-demo env=<env>, creating missing records and updating changed ones in place, keeping them across recreate, and deleting them on down; it must never update or delete a record without that comment, and must refuse when such a record occupies the host name.

## Context

A stable name survives recreating the machine, unlike an sslip.io name, and Let's Encrypt's HTTP-01 check and TLS end on the machine, so the records are not proxied. The comment plays the role of the Hetzner labels: the automation keeps no state and finds its records by it, so it can never delete the docs site's or a hand-made record. Modeled by Claude Code for the demo machine automation. The author asked, through the orchestrating agent, for the demo on the real domain openfief.dev (a Cloudflare zone on the Free plan), fully automated, and for an explicit SSH-off mode because the Hetzner project has no SSH key and Hetzner would otherwise email a root password. The docs site shares the zone (docs.openfief.dev), so the demo must touch only its own records. The author has not reviewed this requirement yet.

## Source

> The author owns the Cloudflare zone openfief.dev and wants the demo on a real domain, fully automated ("remember about automation").

Orchestrator brief for impl/demo-domain; docs/dev/demo.md#dns

> Demo and test environment: a Hetzner machine running k3s, deployed by the pipeline from `main` and from beta releases. It holds synthetic data only: never real broker credentials or portfolio data (GR-12).

docs/dev/sdlc.md#ci-and-environments
