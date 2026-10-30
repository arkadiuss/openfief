---
title: A published release must add each plugins/<name> image to the catalog with the digest it signed, the release version and OpenFief's CI as publisher, refusing a missing digest or a version already listed, on a release commit from which the operator, core and CLI of that release are built.
status: open
text_ref: docs/DESIGN.md#D-88
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-88
  - lane:strict
semantic_text: A published release must add each plugins/<name> image to the catalog with the digest it signed, the release version and OpenFief's CI as publisher, refusing a missing digest or a version already listed, on a release commit from which the operator, core and CLI of that release are built.
logic_claims:
  - CLAIM-4E2FDA7909D37F66
semantic_clauses:
  - A published release must add each plugins/<name> image to the catalog with the digest it signed, the release version and OpenFief's CI as publisher, refusing a missing digest or a version already listed, on a release commit from which the operator, core and CLI of that release are built.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 584411fb650f8059b70262b46ae6b5ad2c76d7e9440d9cbbcb2a461c9b002480
semantic_inventory:
  - claim_key: CLAIM-4E2FDA7909D37F66
    claim_text: A published release must add each plugins/<name> image to the catalog with the digest it signed, the release version and OpenFief's CI as publisher, refusing a missing digest or a version already listed, on a release commit from which the operator, core and CLI of that release are built
    role: normative
    status: modeled
    span:
      start: 0
      end: 287
origin:
  kind: agent
  recorded_at: '2026-10-29T18:02:09.150Z'
id: REQ-tooling-release-catalog-entries
type: req
---
A published release must add each plugins/<name> image to the catalog with the digest it signed, the release version and OpenFief's CI as publisher, refusing a missing digest or a version already listed, on a release commit from which the operator, core and CLI of that release are built.

## Context

D-88 makes the hand-started release add first-party plugins to the catalog. The catalog is embedded in the operator and CLI, so they must be built after the plugins are signed; the release commit and the pull request back to main are this package's reading, recorded in docs/implementation-questions.md. Refines REQ-tooling-release-catalogs-first-party.

## Source

> [D-88] The release workflow adds each first-party plugin to the catalog with the image digest it signed. The author starts every release by hand, and that release is the review D-44 requires, so first-party plugins show as verified.

docs/DESIGN.md#D-88
