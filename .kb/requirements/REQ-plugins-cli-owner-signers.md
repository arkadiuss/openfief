---
title: 'The CLI must trust exactly the signers the operator trusts: OpenFief''s CI, the catalog''s publishers and the owner''s signers from the openfief-trusted-signers ConfigMap (or a file the owner names), and an invalid signer document must stop it.'
status: open
text_ref: docs/DESIGN.md#D-45
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-45
  - lane:strict
semantic_text: 'The CLI must trust exactly the signers the operator trusts: OpenFief''s CI, the catalog''s publishers and the owner''s signers from the openfief-trusted-signers ConfigMap (or a file the owner names), and an invalid signer document must stop it.'
logic_claims:
  - CLAIM-92EF0BDBAE01BB44
semantic_clauses:
  - 'The CLI must trust exactly the signers the operator trusts: OpenFief''s CI, the catalog''s publishers and the owner''s signers from the openfief-trusted-signers ConfigMap (or a file the owner names), and an invalid signer document must stop it.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1f531669066e6dd4bff286ed152e804c7cc9ed715d40b6e388a8026ae1e12ae5
semantic_inventory:
  - claim_key: CLAIM-92EF0BDBAE01BB44
    claim_text: 'The CLI must trust exactly the signers the operator trusts: OpenFief''s CI, the catalog''s publishers and the owner''s signers from the openfief-trusted-signers ConfigMap (or a file the owner names), and an invalid signer document must stop it'
    role: normative
    status: modeled
    span:
      start: 0
      end: 240
origin:
  kind: agent
  recorded_at: '2026-10-28T23:14:13.443Z'
id: REQ-plugins-cli-owner-signers
type: req
---
The CLI must trust exactly the signers the operator trusts: OpenFief's CI, the catalog's publishers and the owner's signers from the openfief-trusted-signers ConfigMap (or a file the owner names), and an invalid signer document must stop it.

## Context

D-56 calls the CLI's check a preview of the operator's; it can only be one if both trust the same signers (D-45). The operator reads the owner's signers from the openfief-trusted-signers ConfigMap, so the CLI reads the same ConfigMap with the owner's kubeconfig; for --print without a cluster the owner may name the file instead. Refines REQ-plugins-cli-install. Modeled by Claude Code for the M0 openfief CLI task (impl/cli); the author has not reviewed it yet.

## Source

> The operator trusts three kinds of signer: OpenFief's CI, the publishers pinned in the catalog, and signers the owner adds.

docs/DESIGN.md D-45
