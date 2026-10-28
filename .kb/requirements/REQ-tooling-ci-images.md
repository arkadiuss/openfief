---
title: CI must build one container image for each Dockerfile at cmd/<name>/Dockerfile or plugins/<name>/Dockerfile, and must skip the image jobs when there is none.
status: open
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
semantic_text: CI must build one container image for each Dockerfile at cmd/<name>/Dockerfile or plugins/<name>/Dockerfile, and must skip the image jobs when there is none.
logic_claims:
  - CLAIM-686AB73E1B8D70C8
semantic_clauses:
  - CI must build one container image for each Dockerfile at cmd/<name>/Dockerfile or plugins/<name>/Dockerfile, and must skip the image jobs when there is none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 27706edce22e11c91e1b0cb49fc1db4c88d3a0b0fdd70662598dd2e170d23d99
semantic_inventory:
  - claim_key: CLAIM-686AB73E1B8D70C8
    claim_text: CI must build one container image for each Dockerfile at cmd/<name>/Dockerfile or plugins/<name>/Dockerfile, and must skip the image jobs when there is none
    role: normative
    status: modeled
    span:
      start: 0
      end: 156
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:30.959Z'
id: REQ-tooling-ci-images
type: req
---
CI must build one container image for each Dockerfile at cmd/<name>/Dockerfile or plugins/<name>/Dockerfile, and must skip the image jobs when there is none.

## Context

docs/dev/sdlc.md asks CI for multi-arch images with an SBOM, signed with cosign (D-17). No Dockerfile exists yet, so discovery keeps the image jobs skipped rather than failing, and later work adds an image by adding a Dockerfile in the layout of docs/dev/code-organization.md. Image names must be unique and lowercase because they become registry paths. Requested in the SDLC/CI implementation brief.

## Source

> CI runs on GitHub Actions **(P)**: the same checks as the hook, end-to-end tests on k3s plus kind (D-12), and multi-arch images with an SBOM, signed with cosign (D-17).

docs/dev/sdlc.md#ci-and-environments
