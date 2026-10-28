---
title: task check must build deploy/demo, every git path its Flux Kustomizations apply and the deploy/demo/openfief template with the pinned kustomize, offline, and must fail when an image is not pinned by a sha256 digest, an OCI source does not require a cosign signature, or a substituted layer uses a variable other than DEMO_HOST and DEMO_ACME_SERVER.
status: open
text_ref: docs/dev/sdlc.md#flow
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-flow
  - lane:strict
semantic_text: task check must build deploy/demo, every git path its Flux Kustomizations apply and the deploy/demo/openfief template with the pinned kustomize, offline, and must fail when an image is not pinned by a sha256 digest, an OCI source does not require a cosign signature, or a substituted layer uses a variable other than DEMO_HOST and DEMO_ACME_SERVER.
logic_claims:
  - CLAIM-3332D66244912FFA
semantic_clauses:
  - task check must build deploy/demo, every git path its Flux Kustomizations apply and the deploy/demo/openfief template with the pinned kustomize, offline, and must fail when an image is not pinned by a sha256 digest, an OCI source does not require a cosign signature, or a substituted layer uses a variable other than DEMO_HOST and DEMO_ACME_SERVER.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5c808a20fec76d1d229ffc387142905c904f3aead6fd14d778e4c56ad2fb8ce9
semantic_inventory:
  - claim_key: CLAIM-3332D66244912FFA
    claim_text: task check must build deploy/demo, every git path its Flux Kustomizations apply and the deploy/demo/openfief template with the pinned kustomize, offline, and must fail when an image is not pinned by a sha256 digest, an OCI source does not require a cosign signature, or a substituted layer uses a variable other than DEMO_HOST and DEMO_ACME_SERVER
    role: normative
    status: modeled
    span:
      start: 0
      end: 347
origin:
  kind: agent
  recorded_at: '2026-10-28T09:01:34.798Z'
id: REQ-tooling-demo-manifests
type: req
---
task check must build deploy/demo, every git path its Flux Kustomizations apply and the deploy/demo/openfief template with the pinned kustomize, offline, and must fail when an image is not pinned by a sha256 digest, an OCI source does not require a cosign signature, or a substituted layer uses a variable other than DEMO_HOST and DEMO_ACME_SERVER.

## Context

Flux applies deploy/demo straight from main, so a manifest that does not build, an unpinned image, an unverified source or a mistyped variable, which Flux would replace with an empty string, would break or weaken the demo without any other check catching it. Modeled by Claude Code for the demo machine automation, which the orchestrating agent specified from the author's request for good committed automation of the demo machine, easy to create and recreate, with spending limited to 10 EUR a month if possible; the author has not reviewed it yet.

## Source

> **Local pipeline.** Tests, lint and vulnerability checks run on the pre-commit hook, and they are the same checks CI runs, so local green means CI green.

docs/dev/sdlc.md#flow
