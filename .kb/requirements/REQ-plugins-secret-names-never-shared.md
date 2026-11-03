---
title: The CLI must refuse an install or an upgrade, prepared or applied, whose grant names a Secret another installed instance names, and the operator must not run a plugin version that would mount a Secret another instance mounts, unless it mounts that Secret too and is the older Plugin.
status: open
text_ref: docs/DESIGN.md#D-89
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-89
  - lane:strict
semantic_text: The CLI must refuse an install or an upgrade, prepared or applied, whose grant names a Secret another installed instance names, and the operator must not run a plugin version that would mount a Secret another instance mounts, unless it mounts that Secret too and is the older Plugin.
logic_claims:
  - CLAIM-3CCC62ACDA78AE97
semantic_clauses:
  - The CLI must refuse an install or an upgrade, prepared or applied, whose grant names a Secret another installed instance names, and the operator must not run a plugin version that would mount a Secret another instance mounts, unless it mounts that Secret too and is the older Plugin.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c310b6cd667bfd7e5d8bc701c6929c261251327ade02b4e210d2522fec42ff5b
semantic_inventory:
  - claim_key: CLAIM-3CCC62ACDA78AE97
    claim_text: The CLI must refuse an install or an upgrade, prepared or applied, whose grant names a Secret another installed instance names, and the operator must not run a plugin version that would mount a Secret another instance mounts, unless it mounts that Secret too and is the older Plugin
    role: exception
    status: modeled
    span:
      start: 0
      end: 282
origin:
  kind: agent
  recorded_at: '2026-11-03T08:22:19.525Z'
id: REQ-plugins-secret-names-never-shared
type: req
---
The CLI must refuse an install or an upgrade, prepared or applied, whose grant names a Secret another installed instance names, and the operator must not run a plugin version that would mount a Secret another instance mounts, unless it mounts that Secret too and is the older Plugin.

## Context

Secret names are <instance>-<role> (D-89), and instance names and roles may both hold hyphens, so instance ibkr with role flex-token and instance ibkr-flex with role token both name ibkr-flex-token. The CLI checked this only at install; an upgrade, or a Plugin and grant applied with kubectl or GitOps, could mount another instance's Secret, such as a broker token. Recorded with default firmness by the implementation agent of the second security review's fixes (fix/review2-r1-r4), from finding R3. Making names unambiguous by construction, such as a separator neither part may hold, would change the name D-89 decides, so it is left to the author; until then the collision is refused everywhere a version is installed, upgraded or run.

## Source

> [D-89] A manifest names each Secret by its role, such as `flex-token`, and the operator mounts the Secret named `<instance>-<role>`, so two instances of one image never share a Secret.

> [D-182] A plugin's Secrets are named so they can never collide with the platform's ...

docs/DESIGN.md#D-89, docs/DESIGN.md#D-182
