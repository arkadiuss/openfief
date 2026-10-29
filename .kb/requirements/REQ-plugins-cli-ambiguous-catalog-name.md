---
title: The CLI must refuse a catalog name that matches several versions and list those versions instead of picking one.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-90-1
  - lane:strict
text_ref: design-decisions:D-90#1
origin:
  kind: import
  ref: design-decisions:D-90#1
  recorded_at: '2026-10-29T08:34:24.588Z'
semantic_text: The CLI must refuse a catalog name that matches several versions and list those versions instead of picking one.
logic_claims:
  - CLAIM-DDFC0ED92D74AAFB
semantic_clauses:
  - The CLI must refuse a catalog name that matches several versions and list those versions instead of picking one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 746809b55403e78368e0ed7eb89542d55c16da1c06b2917d9dcdc0625eeeae15
semantic_inventory:
  - claim_key: CLAIM-DDFC0ED92D74AAFB
    claim_text: The CLI must refuse a catalog name that matches several versions and list those versions instead of picking one
    role: normative
    status: modeled
    span:
      start: 0
      end: 111
id: REQ-plugins-cli-ambiguous-catalog-name
type: req
---
The CLI must refuse a catalog name that matches several versions and list those versions instead of picking one.

## Context

D-44 needs a new catalog entry for each reviewed version, so one plugin name can match several entries, and D-46 has the owner install plugins with the openfief CLI. D-90 has the CLI refuse such a name and list the versions, so the owner always chooses the version deliberately. REQ-plugins-cli-install-source, which the author has not reviewed, already refuses a name several entries share but does not list the versions. The author recorded D-88 to D-90 after the design session that answered the implementation's questions.

## Source

> The CLI refuses a catalog name that matches several versions and lists them instead of picking one, and refuses to install over an existing instance. `openfief plugin upgrade <instance> <version>` verifies the new image, shows how its requested permissions differ, and applies the upgrade only when the owner confirms.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-90#1
