---
title: 'Every release bundle must bring its database: CloudNativePG''s release manifest, refused unless its sha256 is the one pinned in the repository, and a minimal openfief-db cluster the owner can edit.'
status: open
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
semantic_text: 'Every release bundle must bring its database: CloudNativePG''s release manifest, refused unless its sha256 is the one pinned in the repository, and a minimal openfief-db cluster the owner can edit.'
logic_claims:
  - CLAIM-2C3FBB9F7F3BF878
semantic_clauses:
  - 'Every release bundle must bring its database: CloudNativePG''s release manifest, refused unless its sha256 is the one pinned in the repository, and a minimal openfief-db cluster the owner can edit.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ed320a32a7fae72a456927c96dd226daec2770f02e75e1d2e631e45fea38413f
semantic_inventory:
  - claim_key: CLAIM-2C3FBB9F7F3BF878
    claim_text: 'Every release bundle must bring its database: CloudNativePG''s release manifest, refused unless its sha256 is the one pinned in the repository, and a minimal openfief-db cluster the owner can edit'
    role: exception
    status: modeled
    span:
      start: 0
      end: 195
origin:
  kind: agent
  recorded_at: '2026-11-01T21:51:59.686Z'
id: REQ-tooling-release-bundle-database
type: req
---
Every release bundle must bring its database: CloudNativePG's release manifest, refused unless its sha256 is the one pinned in the repository, and a minimal openfief-db cluster the owner can edit.

## Context

The author's release rules in docs/dev/sdlc.md say the bootstrap install brings its database by default (D-14), so an owner doesn't assemble CloudNativePG and the cluster by hand; an owner whose cluster already runs CloudNativePG skips its part. Issue #69 asked for it before the first alpha.

## Source

> The bootstrap install brings its database by default: CloudNativePG's release manifest, pinned by sha256, and a minimal `openfief-db` cluster the owner can edit (D-14). An owner whose cluster already runs CloudNativePG skips its part.

docs/dev/sdlc.md#releases
