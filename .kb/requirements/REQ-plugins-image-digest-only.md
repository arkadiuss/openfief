---
title: A plugin image reference must name a repository and a sha256 digest only, and a reference with a tag, with or without a digest, must be rejected.
status: open
text_ref: docs/DESIGN.md#D-17
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-17
  - lane:strict
semantic_text: A plugin image reference must name a repository and a sha256 digest only, and a reference with a tag, with or without a digest, must be rejected.
logic_claims:
  - CLAIM-0E385B7E3B315936
semantic_clauses:
  - A plugin image reference must name a repository and a sha256 digest only, and a reference with a tag, with or without a digest, must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ffe066c0fe9d3c3523d73f5aee3fcef80af2ad6652a07f8bfe2eb3376a283def
semantic_inventory:
  - claim_key: CLAIM-0E385B7E3B315936
    claim_text: A plugin image reference must name a repository and a sha256 digest only, and a reference with a tag, with or without a digest, must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 144
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:31.410Z'
id: REQ-plugins-image-digest-only
type: req
---
A plugin image reference must name a repository and a sha256 digest only, and a reference with a tag, with or without a digest, must be rejected.

## Context

D-17 requires plugin images to be pinned by digest so that what the owner approved is exactly what runs. A tag can be moved to other content; a reference with both a tag and a digest is ambiguous to readers even though Kubernetes uses the digest, so it is refused. Only sha256 is accepted, the algorithm OCI registries use. Signature verification with cosign is a separate adapter step. Refines REQ-AUTO-AC7E986366466A6A. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> Plugin images must be pinned by digest, and their signatures are verified with cosign.

docs/DESIGN.md D-17
