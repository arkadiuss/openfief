---
title: Uninstalling an instance must delete the Secrets its manifest or grant names, except those another installed instance still names, which it must keep and report.
status: open
text_ref: docs/DESIGN.md#D-58
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-58
  - lane:strict
semantic_text: Uninstalling an instance must delete the Secrets its manifest or grant names, except those another installed instance still names, which it must keep and report.
logic_claims:
  - CLAIM-1CF1C25F56D4B1F1
semantic_clauses:
  - Uninstalling an instance must delete the Secrets its manifest or grant names, except those another installed instance still names, which it must keep and report.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d8743eeca752f44b872454dec3fe546b42d2192763c005b38b3dbc63f1e33ea3
semantic_inventory:
  - claim_key: CLAIM-1CF1C25F56D4B1F1
    claim_text: Uninstalling an instance must delete the Secrets its manifest or grant names, except those another installed instance still names, which it must keep and report
    role: exception
    status: modeled
    span:
      start: 0
      end: 160
origin:
  kind: agent
  recorded_at: '2026-10-28T22:24:17.212Z'
id: REQ-plugins-cli-uninstall-secrets
type: req
---
Uninstalling an instance must delete the Secrets its manifest or grant names, except those another installed instance still names, which it must keep and report.

## Context

D-58 deletes an uninstalled instance's secrets. Secret names come from the image's signed manifest (D-42), so two instances of one image (D-48) name the same Secrets; deleting them would break the instance that stays. The narrowest reading deletes only what no remaining instance names, reported to the author. Refines REQ-plugins-uninstall-keeps-data. Modeled by Claude Code for the M0 openfief CLI task (impl/cli); the author has not reviewed it yet.

## Source

> Uninstalling an instance deletes its workload, grant, secrets and egress allow-list. The data it brought in stays in the ledger (D-18), marked as coming from a removed instance; reinstalling under the same name continues from it.

docs/DESIGN.md D-58
