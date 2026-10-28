---
title: The demo automation must find every resource by the labels app=openfief and env=<name>, create only the resources that are missing, and delete a demo's resources only when CONFIRM equals the demo's name.
status: open
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
semantic_text: The demo automation must find every resource by the labels app=openfief and env=<name>, create only the resources that are missing, and delete a demo's resources only when CONFIRM equals the demo's name.
logic_claims:
  - CLAIM-99130165295D8632
semantic_clauses:
  - The demo automation must find every resource by the labels app=openfief and env=<name>, create only the resources that are missing, and delete a demo's resources only when CONFIRM equals the demo's name.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e867a236d31ec400493825d403b27093a5aaa0483bb1bc0f101818d0f13c0685
semantic_inventory:
  - claim_key: CLAIM-99130165295D8632
    claim_text: The demo automation must find every resource by the labels app=openfief and env=<name>, create only the resources that are missing, and delete a demo's resources only when CONFIRM equals the demo's name
    role: normative
    status: modeled
    span:
      start: 0
      end: 202
origin:
  kind: agent
  recorded_at: '2026-10-28T09:00:25.663Z'
id: REQ-tooling-demo-labels
type: req
---
The demo automation must find every resource by the labels app=openfief and env=<name>, create only the resources that are missing, and delete a demo's resources only when CONFIRM equals the demo's name.

## Context

Agent sessions are ephemeral, so the automation keeps no local state such as an OpenTofu state file; labels in Hetzner are the state, which makes up idempotent and lets several named demos coexist. The confirmation guards against deleting the wrong demo. Modeled by Claude Code for the demo machine automation, which the orchestrating agent specified from the author's request for good committed automation of the demo machine, easy to create and recreate, with spending limited to 10 EUR a month if possible; the author has not reviewed it yet.

## Source

> Demo and test environment: a Hetzner machine running k3s, deployed by the pipeline from `main` and from beta releases. It holds synthetic data only: never real broker credentials or portfolio data (GR-12).

docs/dev/sdlc.md#ci-and-environments
