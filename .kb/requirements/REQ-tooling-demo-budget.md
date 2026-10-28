---
title: Before creating a demo resource, the demo automation must sum the monthly gross price of every resource labelled app=openfief in the Hetzner project plus what it would create, show the arithmetic, and refuse when the total exceeds DEMO_BUDGET_EUR, which defaults to 10 EUR.
status: open
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
semantic_text: Before creating a demo resource, the demo automation must sum the monthly gross price of every resource labelled app=openfief in the Hetzner project plus what it would create, show the arithmetic, and refuse when the total exceeds DEMO_BUDGET_EUR, which defaults to 10 EUR.
logic_claims:
  - CLAIM-B4E640AA32052748
semantic_clauses:
  - Before creating a demo resource, the demo automation must sum the monthly gross price of every resource labelled app=openfief in the Hetzner project plus what it would create, show the arithmetic, and refuse when the total exceeds DEMO_BUDGET_EUR, which defaults to 10 EUR.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bc7415906d0c1625546862e5942d70c97fe9bd24537465f068d1cd2b9a180124
semantic_inventory:
  - claim_key: CLAIM-B4E640AA32052748
    claim_text: Before creating a demo resource, the demo automation must sum the monthly gross price of every resource labelled app=openfief in the Hetzner project plus what it would create, show the arithmetic, and refuse when the total exceeds DEMO_BUDGET_EUR, which defaults to 10 EUR
    role: normative
    status: modeled
    span:
      start: 0
      end: 272
origin:
  kind: agent
  recorded_at: '2026-10-28T09:00:11.530Z'
id: REQ-tooling-demo-budget
type: req
---
Before creating a demo resource, the demo automation must sum the monthly gross price of every resource labelled app=openfief in the Hetzner project plus what it would create, show the arithmetic, and refuse when the total exceeds DEMO_BUDGET_EUR, which defaults to 10 EUR.

## Context

Hetzner Cloud has no spending-limit API, so the budget lives in the scripts; prices come from the live price list at the account's VAT rate and are rounded up in integer micro-euros, never floats. The hard cap is a lower server limit from Hetzner support, documented in docs/dev/demo.md. Modeled by Claude Code for the demo machine automation, which the orchestrating agent specified from the author's request for good committed automation of the demo machine, easy to create and recreate, with spending limited to 10 EUR a month if possible; the author has not reviewed it yet.

## Source

> Demo and test environment: a Hetzner machine running k3s, deployed by the pipeline from `main` and from beta releases. It holds synthetic data only: never real broker credentials or portfolio data (GR-12).

docs/dev/sdlc.md#ci-and-environments
