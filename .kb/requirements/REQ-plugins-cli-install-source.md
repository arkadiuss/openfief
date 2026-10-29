---
title: openfief plugin install must take either the name of exactly one catalog entry or a digest-pinned image reference, must find the catalog entry that pins an image's digest, must refuse a tag, an unknown name or a name several entries share, and must name the instance after the catalog entry or the image repository unless the owner names it.
status: open
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
semantic_text: openfief plugin install must take either the name of exactly one catalog entry or a digest-pinned image reference, must find the catalog entry that pins an image's digest, must refuse a tag, an unknown name or a name several entries share, and must name the instance after the catalog entry or the image repository unless the owner names it.
logic_claims:
  - CLAIM-F0E6462C065C0EAA
semantic_clauses:
  - openfief plugin install must take either the name of exactly one catalog entry or a digest-pinned image reference, must find the catalog entry that pins an image's digest, must refuse a tag, an unknown name or a name several entries share, and must name the instance after the catalog entry or the image repository unless the owner names it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e8246c516b12edcb7e25b054bd7101f200ff7be4934264b935d7a11c15af8594
semantic_inventory:
  - claim_key: CLAIM-F0E6462C065C0EAA
    claim_text: openfief plugin install must take either the name of exactly one catalog entry or a digest-pinned image reference, must find the catalog entry that pins an image's digest, must refuse a tag, an unknown name or a name several entries share, and must name the instance after the catalog entry or the image repository unless the owner names it
    role: exception
    status: modeled
    span:
      start: 0
      end: 340
origin:
  kind: agent
  recorded_at: '2026-10-28T22:24:04.362Z'
id: REQ-plugins-cli-install-source
type: req
---
openfief plugin install must take either the name of exactly one catalog entry or a digest-pinned image reference, must find the catalog entry that pins an image's digest, must refuse a tag, an unknown name or a name several entries share, and must name the instance after the catalog entry or the image repository unless the owner names it.

## Context

D-46 has the CLI read the catalog or the image; D-43 and D-44 make a catalog entry pin a reviewed digest, so an image whose digest the catalog pins is the verified plugin even from a mirror. D-17 refuses tags. Several entries may share a name because each reviewed version needs its own entry (D-44); which one a bare name means is not decided, so the CLI refuses and lists the images (reported to the author). D-48 makes each install a named instance; the default name is the catalog name or the repository's last path component. Refines REQ-plugins-cli-install. Modeled by Claude Code for the M0 openfief CLI task (impl/cli); the author has not reviewed it yet.

## Source

> Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` with the owner's own kubeconfig, or prints them for GitOps. The core has no right to create or change grants; in M0 the UI only shows installed plugins.

docs/DESIGN.md D-46
