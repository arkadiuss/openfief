---
title: Credentials must be stored as Kubernetes Secrets.
status: open
text_ref: design-decisions:D-29#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-29-1
  - lane:strict
semantic_text: Credentials must be stored as Kubernetes Secrets.
logic_claims:
  - CLAIM-BCB16B029C85829C
semantic_clauses:
  - Credentials must be stored as Kubernetes Secrets.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c475113847604e95bfe07fc96b4383d405631a76b02725f5f5278475cec4dee4
semantic_inventory:
  - claim_key: CLAIM-BCB16B029C85829C
    claim_text: Credentials must be stored as Kubernetes Secrets
    role: normative
    status: modeled
    span:
      start: 0
      end: 48
origin:
  kind: import
  ref: design-decisions:D-29#1
  recorded_at: '2026-10-27T20:12:47.886Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-0455DAAE4C01C965
type: req
---
Credentials must be stored as Kubernetes Secrets.

## Source

> Credentials are Kubernetes Secrets, fed by whatever the user already runs (External Secrets Operator, Vault, SOPS and so on). The docs require etcd encryption at rest, encrypted disks and encrypted backups. No home-grown cryptography in M0.

Source: OpenFief author decisions (DESIGN.md section 10) - D-29#1
