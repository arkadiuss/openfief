---
title: A PluginGrant that names no institutions must approve none.
status: open
text_ref: design-decisions:D-185#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-185-3
  - lane:strict
semantic_text: A PluginGrant that names no institutions must approve none.
logic_claims:
  - CLAIM-9C45A1C600F01AC2
semantic_clauses:
  - A PluginGrant that names no institutions must approve none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5b603156c63564616e1256ba6303c7cf67871716fd26ea756ba76eeade9b40d6
semantic_inventory:
  - claim_key: CLAIM-9C45A1C600F01AC2
    claim_text: A PluginGrant that names no institutions must approve none
    role: normative
    status: modeled
    span:
      start: 0
      end: 58
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-185#3
  recorded_at: '2026-11-03T12:32:01.626Z'
id: REQ-plugins-grant-without-institutions-approves-none
type: req
---
A PluginGrant that names no institutions must approve none.

## Why this changed

On 2026-11-03 the author dropped D-185's fallback, under which a grant made before D-185 approved the institutions its plugin already declared. No grant predates D-185, since nothing stored before the first release is kept (D-161), so the fallback only opened a loophole: deleting the field by hand loosened a grant. This requirement supersedes REQ-plugins-legacy-grants-approve-declared-institutions.

## Context

The CLI writes every connector's grant with its institutions, an empty list included, and the demo's and the end-to-end suite's grants name theirs, so every shipped grant keeps booking. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The institutions a connector books for are a permission its grant approves, as its hostnames are (D-16): the signed manifest declares them and the core accepts only what both name. The install warns when another instance already books for an institution, naming that instance. A grant that names no institutions approves none.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-185#3
