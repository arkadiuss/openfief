---
title: Uninstalling a plugin instance must delete its workload, grant, secrets and egress allow-list, and must keep the data it brought in, marked as coming from a removed instance.
status: open
text_ref: design-decisions:D-58#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-58-1
  - lane:strict
semantic_text: Uninstalling a plugin instance must delete its workload, grant, secrets and egress allow-list, and must keep the data it brought in, marked as coming from a removed instance.
logic_claims:
  - CLAIM-2DFAAB34A91C1117
semantic_clauses:
  - Uninstalling a plugin instance must delete its workload, grant, secrets and egress allow-list, and must keep the data it brought in, marked as coming from a removed instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1dad7341ea5dcc206f08f6480c81fb82d021f816502c49431bb93bc4682ddc13
semantic_inventory:
  - claim_key: CLAIM-2DFAAB34A91C1117
    claim_text: Uninstalling a plugin instance must delete its workload, grant, secrets and egress allow-list, and must keep the data it brought in, marked as coming from a removed instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 173
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-58#1
  recorded_at: '2026-10-28T06:09:35.410Z'
id: REQ-plugins-uninstall-keeps-data
type: req
---
Uninstalling a plugin instance must delete its workload, grant, secrets and egress allow-list, and must keep the data it brought in, marked as coming from a removed instance.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> Uninstalling an instance deletes its workload, grant, secrets and egress allow-list. The data it brought in stays in the ledger (D-18), marked as coming from a removed instance; reinstalling under the same name continues from it.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-58#1
