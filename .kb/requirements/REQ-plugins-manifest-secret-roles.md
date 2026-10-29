---
title: A manifest must name each Secret by its role, such as flex-token.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-89-1
  - lane:strict
text_ref: design-decisions:D-89#1
origin:
  kind: import
  ref: design-decisions:D-89#1
  recorded_at: '2026-10-29T08:32:20.983Z'
semantic_text: A manifest must name each Secret by its role, such as flex-token.
logic_claims:
  - CLAIM-FC59B9026762A26D
semantic_clauses:
  - A manifest must name each Secret by its role, such as flex-token.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: a703a3180a637699b184e840bdafee3dd0c9fdfd78d577a59d8bffc4996450a4
semantic_inventory:
  - claim_key: CLAIM-FC59B9026762A26D
    claim_text: A manifest must name each Secret by its role, such as flex-token
    role: normative
    status: modeled
    span:
      start: 0
      end: 64
id: REQ-plugins-manifest-secret-roles
type: req
---
A manifest must name each Secret by its role, such as flex-token.

## Context

D-42 embeds the manifest in the signed image, and D-48 makes each install a separately named instance with its own secrets (REQ-plugins-instance-per-install). A Secret name taken straight from the manifest would be the same for every instance of one image, so D-89 has the manifest name only the Secret's role, the operator mount the Secret named after the instance and the role (REQ-operator-instance-secret-names), and two instances never share a Secret (REQ-plugins-instances-never-share-secrets). The author recorded D-88 to D-90 after the design session that answered the implementation's questions.

## Source

> A manifest names each Secret by its role, such as `flex-token`, and the operator mounts the Secret named `<instance>-<role>`, so two instances of one image never share a Secret.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-89#1
