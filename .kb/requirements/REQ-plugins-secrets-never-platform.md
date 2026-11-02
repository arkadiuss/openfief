---
title: No plugin instance may be named openfief or a name starting with openfief-, no grant may name a Secret starting with openfief-, and uninstalling an instance must delete only the Secrets whose openfief.io/instance label names that instance, read again just before each delete.
status: open
text_ref: docs/DESIGN.md#D-182
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-182
  - lane:strict
semantic_text: No plugin instance may be named openfief or a name starting with openfief-, no grant may name a Secret starting with openfief-, and uninstalling an instance must delete only the Secrets whose openfief.io/instance label names that instance, read again just before each delete.
logic_claims:
  - CLAIM-344A4776C65FF854
semantic_clauses:
  - No plugin instance may be named openfief or a name starting with openfief-, no grant may name a Secret starting with openfief-, and uninstalling an instance must delete only the Secrets whose openfief.io/instance label names that instance, read again just before each delete.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e61af27856653e7996882b44aedd71b66db498057bd6e09e9cdd0275f9a6673c
semantic_inventory:
  - claim_key: CLAIM-344A4776C65FF854
    claim_text: No plugin instance may be named openfief or a name starting with openfief-, no grant may name a Secret starting with openfief-, and uninstalling an instance must delete only the Secrets whose openfief.io/instance label names that instance, read again just before each delete
    role: normative
    status: modeled
    span:
      start: 0
      end: 274
origin:
  kind: agent
  recorded_at: '2026-11-02T22:19:54.406Z'
id: REQ-plugins-secrets-never-platform
type: req
---
No plugin instance may be named openfief or a name starting with openfief-, no grant may name a Secret starting with openfief-, and uninstalling an instance must delete only the Secrets whose openfief.io/instance label names that instance, read again just before each delete.

## Context

Plugin Secrets are named <instance>-<role> (D-89) and the default instance name is the image repository's last segment, so an image named openfief with a role db-app mounted openfief-db-app, the CloudNativePG credentials of the core's database, and its uninstall deleted them. Every object of the platform in the namespace is named openfief or openfief-*, so reserving those names makes a collision impossible. The operator's Role can't read Secrets, so it relies on the reserved names; the CLI, with the owner's kubeconfig, also checks the owner's label before it deletes. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p1-isolation-net), from the security review the author put before the alpha tag. No firm requirement for this part of D-182 existed in kibi on this branch when it was built.

## Source

> A plugin's Secrets are named so they can never collide with the platform's, and a connector books only for the institutions its signed manifest declares.

docs/DESIGN.md#D-182
