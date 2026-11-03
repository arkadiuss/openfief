---
title: For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>--<role>.
status: open
text_ref: design-decisions:D-89#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-89-2
  - lane:strict
semantic_text: For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>--<role>.
logic_claims:
  - CLAIM-A256A8A275B88723
semantic_clauses:
  - For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>--<role>.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cc2cdd46b78e950b43072363d1311fff38756e738e34e7ebaddc19e7842a0c79
semantic_inventory:
  - claim_key: CLAIM-A256A8A275B88723
    claim_text: For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>--<role>
    role: normative
    status: modeled
    span:
      start: 0
      end: 104
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-89#2
  recorded_at: '2026-11-03T18:23:55.726Z'
id: REQ-operator-instance-secret-names-double-hyphen
type: req
---
For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>--<role>.

## Why this changed

On 2026-11-03 the author amended D-89: a Secret named `<instance>-<role>` could be read two ways, since both an instance name and a role may hold hyphens, so instance `ibkr` with role `flex-token` and instance `ibkr-flex` with role `token` both named `ibkr-flex-token`. With `--` as the separator, forbidden inside both parts, a Secret's name reads only one way. REQ-operator-instance-secret-names had the operator mount `<instance>-<role>`, so this requirement replaces it.

## Context

D-11 has the operator compile each manifest into Secret mounts, and D-48 gives each instance its own Secrets. The grant names each Secret in full, so the owner approves exactly the Secret that is mounted. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> [D-89] A manifest names each Secret by its role, such as `flex-token`, and the operator mounts the Secret named `<instance>--<role>`, so two instances of one image never share a Secret. Neither an instance name nor a role may contain `--`, so a Secret's name reads only one way.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-89#2
