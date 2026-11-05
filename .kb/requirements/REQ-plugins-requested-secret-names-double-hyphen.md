---
title: A plugin instance must request, and its workload must mount, the Secret <instance>--<role> for each Secret role its manifest names, and a role must be a DNS label without two hyphens in a row.
status: open
text_ref: docs/DESIGN.md#D-89
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-89
  - lane:strict
semantic_text: A plugin instance must request, and its workload must mount, the Secret <instance>--<role> for each Secret role its manifest names, and a role must be a DNS label without two hyphens in a row.
logic_claims:
  - CLAIM-C97E4F8383CA0D82
semantic_clauses:
  - A plugin instance must request, and its workload must mount, the Secret <instance>--<role> for each Secret role its manifest names, and a role must be a DNS label without two hyphens in a row.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c01a89b81537441368127e37af8400122e37a85e2d0c574d66940c2e310c47d4
semantic_inventory:
  - claim_key: CLAIM-C97E4F8383CA0D82
    claim_text: A plugin instance must request, and its workload must mount, the Secret <instance>--<role> for each Secret role its manifest names, and a role must be a DNS label without two hyphens in a row
    role: normative
    status: modeled
    span:
      start: 0
      end: 191
origin:
  kind: agent
  recorded_at: '2026-11-03T18:24:28.501Z'
id: REQ-plugins-requested-secret-names-double-hyphen
type: req
---
A plugin instance must request, and its workload must mount, the Secret <instance>--<role> for each Secret role its manifest names, and a role must be a DNS label without two hyphens in a row.

## Why this changed

On 2026-11-03 the author amended D-89: a Secret named `<instance>-<role>` could be read two ways, since both an instance name and a role may hold hyphens, so instance `ibkr` with role `flex-token` and instance `ibkr-flex` with role `token` both named `ibkr-flex-token`. With `--` as the separator, forbidden inside both parts, a Secret's name reads only one way. REQ-plugins-requested-secret-names said `<instance>-<role>`, so this requirement replaces it.

## Context

D-89 names Secrets by role in the manifest. The grant names the Secret in full so the owner approves exactly what is mounted; a DNS-label role keeps the name a valid Secret name. Refines REQ-plugins-manifest-secret-roles and REQ-operator-instance-secret-names-double-hyphen.

## Source

> [D-89] A manifest names each Secret by its role, such as `flex-token`, and the operator mounts the Secret named `<instance>--<role>`, so two instances of one image never share a Secret. Neither an instance name nor a role may contain `--`, so a Secret's name reads only one way.

docs/DESIGN.md#D-89
