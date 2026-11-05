---
title: Neither a plugin instance name nor a Secret role may contain two hyphens in a row (--).
status: open
text_ref: design-decisions:D-89#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-89-4
  - lane:strict
semantic_text: Neither a plugin instance name nor a Secret role may contain two hyphens in a row (--).
logic_claims:
  - CLAIM-449BC13E660D55AA
semantic_clauses:
  - Neither a plugin instance name nor a Secret role may contain two hyphens in a row (--).
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a8a43043f2769a3a869648d32404d33f92281d659f31016089d133c5be76a41d
semantic_inventory:
  - claim_key: CLAIM-449BC13E660D55AA
    claim_text: Neither a plugin instance name nor a Secret role may contain two hyphens in a row (--)
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 86
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-89#4
  recorded_at: '2026-11-03T18:24:03.328Z'
id: REQ-plugins-no-double-hyphen-in-instance-or-role
type: req
---
Neither a plugin instance name nor a Secret role may contain two hyphens in a row (--).

## Context

On 2026-11-03 the author amended D-89: a Secret named `<instance>-<role>` could be read two ways, since both an instance name and a role may hold hyphens, so instance `ibkr` with role `flex-token` and instance `ibkr-flex` with role `token` both named `ibkr-flex-token`. With `--` as the separator, forbidden inside both parts, a Secret's name reads only one way. The CRDs, the domain's validation and the CLI's default instance name all refuse such a name or role. The earlier collision checks (REQ-plugins-secret-names-never-shared) stay as a second layer. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> [D-89] A manifest names each Secret by its role, such as `flex-token`, and the operator mounts the Secret named `<instance>--<role>`, so two instances of one image never share a Secret. Neither an instance name nor a role may contain `--`, so a Secret's name reads only one way.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-89#4
