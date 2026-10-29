---
title: openfief plugin upgrade must change nothing unless the owner confirms, and on confirmation replace the Plugin's spec before its grant's, granting the operator read access to a new in-cluster service first and taking it back from one no longer named.
status: open
text_ref: docs/DESIGN.md#D-90
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-90
  - lane:strict
semantic_text: openfief plugin upgrade must change nothing unless the owner confirms, and on confirmation replace the Plugin's spec before its grant's, granting the operator read access to a new in-cluster service first and taking it back from one no longer named.
logic_claims:
  - CLAIM-C9034985B4EB49BC
semantic_clauses:
  - openfief plugin upgrade must change nothing unless the owner confirms, and on confirmation replace the Plugin's spec before its grant's, granting the operator read access to a new in-cluster service first and taking it back from one no longer named.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d8dce04f78f75caa5d52704dfeb4a69d3aeb69990d4e768e23259705fdbf1104
semantic_inventory:
  - claim_key: CLAIM-C9034985B4EB49BC
    claim_text: openfief plugin upgrade must change nothing unless the owner confirms, and on confirmation replace the Plugin's spec before its grant's, granting the operator read access to a new in-cluster service first and taking it back from one no longer named
    role: exception
    status: modeled
    span:
      start: 0
      end: 248
origin:
  kind: agent
  recorded_at: '2026-10-29T17:04:21.422Z'
id: REQ-plugins-upgrade-applies-on-confirmation
type: req
---
openfief plugin upgrade must change nothing unless the owner confirms, and on confirmation replace the Plugin's spec before its grant's, granting the operator read access to a new in-cluster service first and taking it back from one no longer named.

## Context

D-90 applies the upgrade only when the owner confirms; D-16 keeps the previous version running until the grant approves the new one, which the Plugin-then-grant order uses. Refines REQ-plugins-cli-upgrade.

## Source

> [D-90] The CLI refuses a catalog name that matches several versions and lists them instead of picking one, and refuses to install over an existing instance. `openfief plugin upgrade <instance> <version>` verifies the new image, shows how its requested permissions differ, and applies the upgrade only when the owner confirms.

docs/DESIGN.md#D-90
