---
title: openfief plugin install must refuse an instance whose grant names a Secret that another installed instance already names.
status: open
text_ref: docs/DESIGN.md#D-89
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-89
  - lane:strict
semantic_text: openfief plugin install must refuse an instance whose grant names a Secret that another installed instance already names.
logic_claims:
  - CLAIM-10DDE6FB79500EBC
semantic_clauses:
  - openfief plugin install must refuse an instance whose grant names a Secret that another installed instance already names.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8ac230fe7d883f2d8c60ca63d562c48a4e40db903fad621806bbf140c3bbea05
semantic_inventory:
  - claim_key: CLAIM-10DDE6FB79500EBC
    claim_text: openfief plugin install must refuse an instance whose grant names a Secret that another installed instance already names
    role: normative
    status: modeled
    span:
      start: 0
      end: 120
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:44.122Z'
id: REQ-plugins-cli-install-secret-in-use
type: req
---
openfief plugin install must refuse an instance whose grant names a Secret that another installed instance already names.

## Context

D-89 says two instances of one image never share a Secret. Role names are joined to the instance name with a hyphen, so instance a-b with role c and instance a with role b-c spell the same Secret; the CLI refuses such an install rather than let two instances share one. Refines REQ-plugins-instances-never-share-secrets.

## Source

> [D-89] A manifest names each Secret by its role, such as `flex-token`, and the operator mounts the Secret named `<instance>-<role>`, so two instances of one image never share a Secret.

docs/DESIGN.md#D-89
