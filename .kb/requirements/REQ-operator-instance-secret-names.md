---
title: For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>-<role>.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-89-2
  - lane:strict
text_ref: design-decisions:D-89#2
origin:
  kind: import
  ref: design-decisions:D-89#2
  recorded_at: '2026-10-29T08:32:58.034Z'
semantic_text: For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>-<role>.
logic_claims:
  - CLAIM-46766ECC6EC662A2
semantic_clauses:
  - For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>-<role>.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: af3c00445212d116fc774f5c95c9a1b4457a9c1cdd9e5bf47075cb4ae1d41035
semantic_inventory:
  - claim_key: CLAIM-46766ECC6EC662A2
    claim_text: For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>-<role>
    role: normative
    status: modeled
    span:
      start: 0
      end: 103
id: REQ-operator-instance-secret-names
type: req
---
For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>-<role>.

## Context

D-11 has the operator compile each manifest into Secret mounts (REQ-AUTO-761F6500ECFC0316), and D-48 gives each instance its own secrets. D-89 has the manifest name each Secret by its role (REQ-plugins-manifest-secret-roles) and the operator mount the Secret named after the instance and that role, so two instances of one image never share a Secret (REQ-plugins-instances-never-share-secrets). REQ-operator-workload-objects, REQ-plugins-workload-identity-secrets and REQ-plugins-cli-uninstall-secrets, which the author has not reviewed, still treat the manifest's names as the Secrets' names. The author recorded D-88 to D-90 after the design session that answered the implementation's questions.

## Source

> A manifest names each Secret by its role, such as `flex-token`, and the operator mounts the Secret named `<instance>-<role>`, so two instances of one image never share a Secret.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-89#2
