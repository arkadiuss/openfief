---
title: Where the read API lists a transaction, it must name the connector instance that brought it in and mark it as coming from a removed instance when no Plugin of that name is installed, deriving the mark when read and marking nothing while the installed plugins are unknown.
status: open
text_ref: docs/DESIGN.md#D-58
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-58
  - lane:strict
semantic_text: Where the read API lists a transaction, it must name the connector instance that brought it in and mark it as coming from a removed instance when no Plugin of that name is installed, deriving the mark when read and marking nothing while the installed plugins are unknown.
logic_claims:
  - CLAIM-0D14417AFD55E0E4
semantic_clauses:
  - Where the read API lists a transaction, it must name the connector instance that brought it in and mark it as coming from a removed instance when no Plugin of that name is installed, deriving the mark when read and marking nothing while the installed plugins are unknown.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 80156927b55178206a15dc9a6c5270d50539702a2db68331d09f7aadaaaec8ea
semantic_inventory:
  - claim_key: CLAIM-0D14417AFD55E0E4
    claim_text: Where the read API lists a transaction, it must name the connector instance that brought it in and mark it as coming from a removed instance when no Plugin of that name is installed, deriving the mark when read and marking nothing while the installed plugins are unknown
    role: normative
    status: modeled
    span:
      start: 0
      end: 270
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:52.802Z'
id: REQ-views-removed-instance-mark
type: req
---
Where the read API lists a transaction, it must name the connector instance that brought it in and mark it as coming from a removed instance when no Plugin of that name is installed, deriving the mark when read and marking nothing while the installed plugins are unknown.

## Context

D-58: the data a removed instance brought in stays, marked as coming from a removed instance; reinstalling under the same name continues from it. Deriving the mark at read time from the plugin list keeps the append-only log untouched and makes a reinstall clear it. Modeled by Claude Code for the plugin status task; the author has not reviewed it yet.

## Source

> **[D-58]** Uninstalling an instance deletes its workload, grant, secrets and egress allow-list. The data it brought in stays in the ledger (D-18), marked as coming from a removed instance; reinstalling under the same name continues from it.

docs/DESIGN.md#D-58
