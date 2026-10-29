---
title: openfief plugin upgrade must read its version as a digest-pinned image reference or as the version of a catalog entry of the instance's own plugin, found by its current digest or else its repository, and refuse a version no such entry has or several share.
status: open
text_ref: docs/DESIGN.md#D-90
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-90
  - lane:strict
semantic_text: openfief plugin upgrade must read its version as a digest-pinned image reference or as the version of a catalog entry of the instance's own plugin, found by its current digest or else its repository, and refuse a version no such entry has or several share.
logic_claims:
  - CLAIM-3A57F0D3E2BDAB6C
semantic_clauses:
  - openfief plugin upgrade must read its version as a digest-pinned image reference or as the version of a catalog entry of the instance's own plugin, found by its current digest or else its repository, and refuse a version no such entry has or several share.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1532a9bd285a85c7d280a19a2c0fbd478f06b7a1245853f7228a7dbb29f3aefd
semantic_inventory:
  - claim_key: CLAIM-3A57F0D3E2BDAB6C
    claim_text: openfief plugin upgrade must read its version as a digest-pinned image reference or as the version of a catalog entry of the instance's own plugin, found by its current digest or else its repository, and refuse a version no such entry has or several share
    role: normative
    status: modeled
    span:
      start: 0
      end: 255
origin:
  kind: agent
  recorded_at: '2026-10-29T17:04:09.797Z'
id: REQ-plugins-upgrade-version-resolution
type: req
---
openfief plugin upgrade must read its version as a digest-pinned image reference or as the version of a catalog entry of the instance's own plugin, found by its current digest or else its repository, and refuse a version no such entry has or several share.

## Context

D-90 names the command's argument a version but not how a version is written. Catalog entries carry the release version that added them (D-88), so a version picks an entry of the plugin the instance already runs; an image reference covers plugins outside the catalog. Refines REQ-plugins-cli-upgrade.

## Source

> [D-90] The CLI refuses a catalog name that matches several versions and lists them instead of picking one, and refuses to install over an existing instance. `openfief plugin upgrade <instance> <version>` verifies the new image, shows how its requested permissions differ, and applies the upgrade only when the owner confirms.

docs/DESIGN.md#D-90
