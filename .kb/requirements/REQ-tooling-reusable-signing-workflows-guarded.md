---
title: Every reusable workflow that signs keylessly, images.yml and deploy-demo.yml among them, must run its signing jobs only when github.repository is arkadiuss/openfief.
status: open
text_ref: docs/DESIGN.md#D-181
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-181
  - lane:strict
semantic_text: Every reusable workflow that signs keylessly, images.yml and deploy-demo.yml among them, must run its signing jobs only when github.repository is arkadiuss/openfief.
logic_claims:
  - CLAIM-09C32C90692FECFC
semantic_clauses:
  - Every reusable workflow that signs keylessly, images.yml and deploy-demo.yml among them, must run its signing jobs only when github.repository is arkadiuss/openfief.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d980771810ee2aa9a75731c83e9bd6dcfe53d7b4e93eeef34303c18d4320da7b
semantic_inventory:
  - claim_key: CLAIM-09C32C90692FECFC
    claim_text: Every reusable workflow that signs keylessly, images.yml and deploy-demo.yml among them, must run its signing jobs only when github.repository is arkadiuss/openfief
    role: normative
    status: modeled
    span:
      start: 0
      end: 164
origin:
  kind: agent
  recorded_at: '2026-11-03T02:14:38.029Z'
id: REQ-tooling-reusable-signing-workflows-guarded
type: req
---
Every reusable workflow that signs keylessly, images.yml and deploy-demo.yml among them, must run its signing jobs only when github.repository is arkadiuss/openfief.

## Context

D-181 pins OpenFief's CI signer to the source repository and ref of the signing run because a reusable workflow can be called from any repository. Verifiers already pin the repository and ref (REQ-tooling-verify-pins-source-repository, REQ-plugins-signer-source-pinned), but Flux's matchOIDCIdentity on the demo can't, so deploy-demo.yml's signature relies on Flux pulling only from this repository's GHCR package. Refusing to sign in a foreign repository at the workflow itself is a second layer: in a called workflow github.repository is the caller's repository, so the job is skipped before an OIDC token is requested. Modeled by Claude Code for the last fixes before the alpha tag, as the orchestrator asked on 2026-11-03.

## Source

> D-181: OpenFief's CI signer (D-45) is pinned to the source repository and ref of the signing run, arkadiuss/openfief at refs/heads/main, not only to the workflow named in the certificate, because a reusable workflow can be called from any repository.

docs/DESIGN.md D-181
