---
title: A signature must count as a trusted signer's only when its certificate has the signer's exact OIDC issuer and either the signer's exact subject or a subject that the signer's pattern matches in full, or when the signer's public key verified it.
status: open
text_ref: docs/DESIGN.md#D-45
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-45
  - lane:strict
semantic_text: A signature must count as a trusted signer's only when its certificate has the signer's exact OIDC issuer and either the signer's exact subject or a subject that the signer's pattern matches in full, or when the signer's public key verified it.
logic_claims:
  - CLAIM-37A68FC72DBF9EAF
semantic_clauses:
  - A signature must count as a trusted signer's only when its certificate has the signer's exact OIDC issuer and either the signer's exact subject or a subject that the signer's pattern matches in full, or when the signer's public key verified it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 89a85a5303f6fb93194780b9b2fe11c8206020f70ca419386649f5d9d30d8eb7
semantic_inventory:
  - claim_key: CLAIM-37A68FC72DBF9EAF
    claim_text: A signature must count as a trusted signer's only when its certificate has the signer's exact OIDC issuer and either the signer's exact subject or a subject that the signer's pattern matches in full, or when the signer's public key verified it
    role: normative
    status: modeled
    span:
      start: 0
      end: 243
origin:
  kind: agent
  recorded_at: '2026-10-28T08:51:25.780Z'
id: REQ-plugins-signer-identity-match
type: req
---
A signature must count as a trusted signer's only when its certificate has the signer's exact OIDC issuer and either the signer's exact subject or a subject that the signer's pattern matches in full, or when the signer's public key verified it.

## Context

D-45 names three kinds of trusted signer: OpenFief's CI, catalog publishers and signers the owner adds, keyless (their own CI identity) or with a local key. OpenFief's CI identity is the reusable images.yml workflow at refs/heads/main, which ci.yml and release.yml call and which signs only from main; a fork or another branch must not count. Patterns are anchored at both ends so a prefix or suffix cannot sneak through. Refines REQ-plugins-trusted-signers-only. Modeled by Claude Code for the M0 signature-verification task; the author has not reviewed it yet.

## Source

> The operator trusts three kinds of signer: OpenFief's CI, the publishers pinned in the catalog, and signers the owner adds. Owners sign their own plugins with their own CI identity or a local key, and those plugins show as unverified.

docs/DESIGN.md D-45
