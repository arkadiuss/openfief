---
title: openfief plugin install must verify the image with the same trust policy and signature check as the operator before showing or applying anything, must refuse an image that policy refuses, a config the manifest's schema rejects and a name that is already installed, and must apply nothing until the owner confirms.
status: open
text_ref: docs/DESIGN.md#D-56
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-56
  - lane:strict
semantic_text: openfief plugin install must verify the image with the same trust policy and signature check as the operator before showing or applying anything, must refuse an image that policy refuses, a config the manifest's schema rejects and a name that is already installed, and must apply nothing until the owner confirms.
logic_claims:
  - CLAIM-1BFFA36586D94C62
semantic_clauses:
  - openfief plugin install must verify the image with the same trust policy and signature check as the operator before showing or applying anything, must refuse an image that policy refuses, a config the manifest's schema rejects and a name that is already installed, and must apply nothing until the owner confirms.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e7a5a447431158becd633d7f280523fa6eded1d7dfa90e378ff6b565efb9ead8
semantic_inventory:
  - claim_key: CLAIM-1BFFA36586D94C62
    claim_text: openfief plugin install must verify the image with the same trust policy and signature check as the operator before showing or applying anything, must refuse an image that policy refuses, a config the manifest's schema rejects and a name that is already installed, and must apply nothing until the owner confirms
    role: normative
    status: modeled
    span:
      start: 0
      end: 312
origin:
  kind: agent
  recorded_at: '2026-10-28T22:43:32.207Z'
id: REQ-plugins-cli-install-verifies
type: req
---
openfief plugin install must verify the image with the same trust policy and signature check as the operator before showing or applying anything, must refuse an image that policy refuses, a config the manifest's schema rejects and a name that is already installed, and must apply nothing until the owner confirms.

## Context

D-46 has the CLI verify the signature; D-56 calls the CLI's check a preview of the operator's own, and D-57 never deploys an image without a trusted signature, so installing one would only leave a Plugin that never runs. The CLI reuses the operator's cosign adapter and plugins.TrustPolicy, with the owner's signers from the cluster's openfief-trusted-signers ConfigMap or a file. Installing over an existing name would be an upgrade, which D-47 leaves to the owner's own edit. Refines REQ-plugins-cli-install. Modeled by Claude Code for the M0 openfief CLI task (impl/cli); the author has not reviewed it yet.

## Source

> Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` with the owner's own kubeconfig, or prints them for GitOps. The core has no right to create or change grants; in M0 the UI only shows installed plugins.

docs/DESIGN.md D-46
