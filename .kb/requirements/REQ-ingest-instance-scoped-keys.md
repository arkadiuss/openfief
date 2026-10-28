---
title: The core must prefix the key source of every transaction a connector instance delivers with plugin:<instance>:, so no connector can write or supersede another instance's or an in-core connector's transactions.
status: open
text_ref: docs/DESIGN.md#D-48
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-48
  - lane:strict
semantic_text: The core must prefix the key source of every transaction a connector instance delivers with plugin:<instance>:, so no connector can write or supersede another instance's or an in-core connector's transactions.
logic_claims:
  - CLAIM-AF3E361081306D21
semantic_clauses:
  - The core must prefix the key source of every transaction a connector instance delivers with plugin:<instance>:, so no connector can write or supersede another instance's or an in-core connector's transactions.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 67159e7dd4265811d7d6c7bedac9646950c9a8f389e370e3db973e061c9d851c
semantic_inventory:
  - claim_key: CLAIM-AF3E361081306D21
    claim_text: The core must prefix the key source of every transaction a connector instance delivers with plugin:<instance>:, so no connector can write or supersede another instance's or an in-core connector's transactions
    role: normative
    status: modeled
    span:
      start: 0
      end: 208
origin:
  kind: agent
  recorded_at: '2026-10-28T08:55:40.609Z'
id: REQ-ingest-instance-scoped-keys
type: req
---
The core must prefix the key source of every transaction a connector instance delivers with plugin:<instance>:, so no connector can write or supersede another instance's or an in-core connector's transactions.

## Context

D-18 keys transactions by source plus the institution's own ID; D-48 makes each install a separately named instance, and D-58 lets an instance reinstalled under the same name continue from its data. Scoping by instance name gives both. Refines REQ-plugins-instance-per-install. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> Each install is a separately named instance with its own config, secrets, grant and egress allow-list.

docs/DESIGN.md D-48
