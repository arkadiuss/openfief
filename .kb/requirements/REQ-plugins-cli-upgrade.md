---
title: openfief plugin upgrade <instance> <version> must verify the new image, show how its requested permissions differ, and apply the upgrade only when the owner confirms.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-90-3
  - lane:strict
text_ref: design-decisions:D-90#3
origin:
  kind: import
  ref: design-decisions:D-90#3
  recorded_at: '2026-10-29T08:35:34.026Z'
semantic_text: openfief plugin upgrade <instance> <version> must verify the new image, show how its requested permissions differ, and apply the upgrade only when the owner confirms.
logic_claims:
  - CLAIM-8494FD797DB3779D
semantic_clauses:
  - openfief plugin upgrade <instance> <version> must verify the new image, show how its requested permissions differ, and apply the upgrade only when the owner confirms.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 125b44d8a9e22556c20ad400ed3d8cdec86c09e54b4d0ac22f4f3bdecae60607
semantic_inventory:
  - claim_key: CLAIM-8494FD797DB3779D
    claim_text: openfief plugin upgrade <instance> <version> must verify the new image, show how its requested permissions differ, and apply the upgrade only when the owner confirms
    role: normative
    status: modeled
    span:
      start: 0
      end: 165
id: REQ-plugins-cli-upgrade
type: req
---
openfief plugin upgrade <instance> <version> must verify the new image, show how its requested permissions differ, and apply the upgrade only when the owner confirms.

## Context

D-47 keeps plugin upgrades manual even when the permissions are unchanged (REQ-plugins-manual-upgrades), and D-16 keeps a version that requests more permissions from deploying until they are approved (REQ-AUTO-8EA5542EFA92FFCD). D-90 gives the owner the command for it: verify the new image as an install does (D-46), show how its requested permissions differ from the running version's, and apply nothing until the owner confirms. The author recorded D-88 to D-90 after the design session that answered the implementation's questions.

## Source

> The CLI refuses a catalog name that matches several versions and lists them instead of picking one, and refuses to install over an existing instance. `openfief plugin upgrade <instance> <version>` verifies the new image, shows how its requested permissions differ, and applies the upgrade only when the owner confirms.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-90#3
