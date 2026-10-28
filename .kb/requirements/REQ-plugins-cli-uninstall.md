---
title: openfief plugin uninstall must delete the instance's PluginGrant, its Plugin, whose workload the cluster garbage-collects and whose egress allow-list entry the operator drops, and the Secrets no other instance names, and must never touch the data the instance brought into the ledger.
status: open
text_ref: docs/DESIGN.md#D-58
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-58
  - lane:strict
semantic_text: openfief plugin uninstall must delete the instance's PluginGrant, its Plugin, whose workload the cluster garbage-collects and whose egress allow-list entry the operator drops, and the Secrets no other instance names, and must never touch the data the instance brought into the ledger.
logic_claims:
  - CLAIM-013BF51FDA453CC8
semantic_clauses:
  - openfief plugin uninstall must delete the instance's PluginGrant, its Plugin, whose workload the cluster garbage-collects and whose egress allow-list entry the operator drops, and the Secrets no other instance names, and must never touch the data the instance brought into the ledger.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 96607766d6f82691039f34fff20767c5b7d3ec8c48126c60a5fd50b8c1bacf7a
semantic_inventory:
  - claim_key: CLAIM-013BF51FDA453CC8
    claim_text: openfief plugin uninstall must delete the instance's PluginGrant, its Plugin, whose workload the cluster garbage-collects and whose egress allow-list entry the operator drops, and the Secrets no other instance names, and must never touch the data the instance brought into the ledger
    role: normative
    status: modeled
    span:
      start: 0
      end: 283
origin:
  kind: agent
  recorded_at: '2026-10-28T22:43:37.343Z'
id: REQ-plugins-cli-uninstall
type: req
---
openfief plugin uninstall must delete the instance's PluginGrant, its Plugin, whose workload the cluster garbage-collects and whose egress allow-list entry the operator drops, and the Secrets no other instance names, and must never touch the data the instance brought into the ledger.

## Context

D-58 deletes an uninstalled instance's workload, grant, secrets and egress allow-list and keeps its data. The operator makes the Plugin the owner of every workload object, so deleting the Plugin lets garbage collection remove them, and its reconcile of a deleted Plugin drops the allow-list key; the CLI deletes only what nothing else removes. Marking the data as coming from a removed instance is the core's job and not built (reported to the author). Refines REQ-plugins-uninstall-keeps-data. Modeled by Claude Code for the M0 openfief CLI task (impl/cli); the author has not reviewed it yet.

## Source

> Uninstalling an instance deletes its workload, grant, secrets and egress allow-list. The data it brought in stays in the ledger (D-18), marked as coming from a removed instance; reinstalling under the same name continues from it.

docs/DESIGN.md D-58
