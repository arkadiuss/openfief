---
title: Every push to main and every beta must deploy to the demo without a manual step, by publishing deploy/demo/openfief pinned to the image digests whose keyless cosign signatures from this repository's images workflow on main verify, as an OCI artifact signed by the deploy workflow, which Flux applies only after verifying that signature.
status: open
text_ref: docs/DESIGN.md#D-17
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-17
  - lane:strict
semantic_text: Every push to main and every beta must deploy to the demo without a manual step, by publishing deploy/demo/openfief pinned to the image digests whose keyless cosign signatures from this repository's images workflow on main verify, as an OCI artifact signed by the deploy workflow, which Flux applies only after verifying that signature.
logic_claims:
  - CLAIM-2BED11A9401784F5
semantic_clauses:
  - Every push to main and every beta must deploy to the demo without a manual step, by publishing deploy/demo/openfief pinned to the image digests whose keyless cosign signatures from this repository's images workflow on main verify, as an OCI artifact signed by the deploy workflow, which Flux applies only after verifying that signature.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 86c8c4f31223247b9f4b0f47000cd91d0af6f4eb8479d4dfca6effed9e0d8b10
semantic_inventory:
  - claim_key: CLAIM-2BED11A9401784F5
    claim_text: Every push to main and every beta must deploy to the demo without a manual step, by publishing deploy/demo/openfief pinned to the image digests whose keyless cosign signatures from this repository's images workflow on main verify, as an OCI artifact signed by the deploy workflow, which Flux applies only after verifying that signature
    role: normative
    status: modeled
    span:
      start: 0
      end: 335
origin:
  kind: agent
  recorded_at: '2026-10-28T09:01:24.601Z'
id: REQ-tooling-demo-signed-deploy
type: req
---
Every push to main and every beta must deploy to the demo without a manual step, by publishing deploy/demo/openfief pinned to the image digests whose keyless cosign signatures from this repository's images workflow on main verify, as an OCI artifact signed by the deploy workflow, which Flux applies only after verifying that signature.

## Context

D-17 pins images by digest and verifies their cosign signatures; D-57 deploys nothing without a valid signature from a trusted signer. The author asked that a merge to main shows up on the demo without manual steps; publishing a signed artifact that Flux pulls needs no commit to main and no cluster credential outside the machine, unlike the previous kubeconfig-based job. Modeled by Claude Code for the demo machine automation, which the orchestrating agent specified from the author's request for good committed automation of the demo machine, easy to create and recreate, with spending limited to 10 EUR a month if possible; the author has not reviewed it yet.

## Source

> Plugin images must be pinned by digest, and their signatures are verified with cosign.

docs/DESIGN.md D-17

> Demo and test environment: a Hetzner machine running k3s, deployed by the pipeline from `main` and from beta releases. It holds synthetic data only: never real broker credentials or portfolio data (GR-12).

docs/dev/sdlc.md#ci-and-environments
