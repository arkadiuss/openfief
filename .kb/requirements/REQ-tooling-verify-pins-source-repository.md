---
title: Every cosign verification of OpenFief's CLI files or images, in the install guide and in the release, image and demo scripts, must pin the signing run's repository and ref with --certificate-github-workflow-repository and --certificate-github-workflow-ref besides the certificate identity and issuer.
status: open
text_ref: docs/DESIGN.md#D-181
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-181
  - lane:strict
semantic_text: Every cosign verification of OpenFief's CLI files or images, in the install guide and in the release, image and demo scripts, must pin the signing run's repository and ref with --certificate-github-workflow-repository and --certificate-github-workflow-ref besides the certificate identity and issuer.
logic_claims:
  - CLAIM-CF21E49F98B83EB6
semantic_clauses:
  - Every cosign verification of OpenFief's CLI files or images, in the install guide and in the release, image and demo scripts, must pin the signing run's repository and ref with --certificate-github-workflow-repository and --certificate-github-workflow-ref besides the certificate identity and issuer.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d4e7499ea20480e2c631be325b5c5f1632d706defd89b58cf2ca6d3bda2b906e
semantic_inventory:
  - claim_key: CLAIM-CF21E49F98B83EB6
    claim_text: Every cosign verification of OpenFief's CLI files or images, in the install guide and in the release, image and demo scripts, must pin the signing run's repository and ref with --certificate-github-workflow-repository and --certificate-github-workflow-ref besides the certificate identity and issuer
    role: normative
    status: modeled
    span:
      start: 0
      end: 299
origin:
  kind: agent
  recorded_at: '2026-11-02T20:00:27.427Z'
id: REQ-tooling-verify-pins-source-repository
type: req
---
Every cosign verification of OpenFief's CLI files or images, in the install guide and in the release, image and demo scripts, must pin the signing run's repository and ref with --certificate-github-workflow-repository and --certificate-github-workflow-ref besides the certificate identity and issuer.

## Context

The same finding applies to verification outside the operator: images.yml is a reusable workflow, so its certificate identity alone accepts another repository's run that calls it, and the install guide is what the owner runs before trusting a release. The CLI's release.yml is not reusable, but pinning the run's repository and ref costs nothing and keeps one rule. cosign checks these flags against Fulcio's GitHub workflow repository and ref extensions. The author decided that the install guide's cosign commands pin the same repository and ref as the CI signer (D-181). Refines REQ-tooling-release-verify-signatures. Modeled by Claude Code for the P0 security fixes before the alpha release, from the security review's finding that OpenFief's CI identity could be obtained by any repository; the author decided the fix as D-181.

## Source

> D-181: the install guide's cosign commands pin the same source repository and ref.

docs/DESIGN.md D-181 (author decision on the security review's finding P0-1), GR-12
