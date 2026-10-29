---
title: Two instances of one image must never share a Secret.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-89-3
  - lane:strict
text_ref: design-decisions:D-89#3
origin:
  kind: import
  ref: design-decisions:D-89#3
  recorded_at: '2026-10-29T08:33:37.875Z'
semantic_text: Two instances of one image must never share a Secret.
logic_claims:
  - CLAIM-E2BB8D0BD16FB5CD
semantic_clauses:
  - Two instances of one image must never share a Secret.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2edc2e758d8644fcfd7fa272014dee6005ad864a911b152245961b947b8c3b2d
semantic_inventory:
  - claim_key: CLAIM-E2BB8D0BD16FB5CD
    claim_text: Two instances of one image must never share a Secret
    role: normative
    status: modeled
    span:
      start: 0
      end: 52
id: REQ-plugins-instances-never-share-secrets
type: req
---
Two instances of one image must never share a Secret.

## Context

D-48 lets one image be installed more than once, such as one instance per IBKR account, each with its own secrets. D-89 names Secrets by role in the manifest (REQ-plugins-manifest-secret-roles) and after the instance in the cluster (REQ-operator-instance-secret-names), so that two instances of one image never share a Secret. REQ-plugins-cli-uninstall-secrets, which the author has not reviewed, still keeps a Secret another instance names on uninstall, which assumes instances can share one. The author recorded D-88 to D-90 after the design session that answered the implementation's questions.

## Source

> A manifest names each Secret by its role, such as `flex-token`, and the operator mounts the Secret named `<instance>-<role>`, so two instances of one image never share a Secret.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-89#3
