---
title: The Inbox's open view must offer Ask suggestions about its open items in a side column, each answered on the overview.
status: open
text_ref: docs/DESIGN.md#D-194
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-194
  - lane:strict
semantic_text: The Inbox's open view must offer Ask suggestions about its open items in a side column, each answered on the overview.
logic_claims:
  - CLAIM-9FE86209E7403EFF
semantic_clauses:
  - The Inbox's open view must offer Ask suggestions about its open items in a side column, each answered on the overview.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: af2b33a4b574f16b73393ea0a9b80d3f2d35810431955f2dc90997a18d15dcb5
semantic_inventory:
  - claim_key: CLAIM-9FE86209E7403EFF
    claim_text: The Inbox's open view must offer Ask suggestions about its open items in a side column, each answered on the overview
    role: normative
    status: modeled
    span:
      start: 0
      end: 117
origin:
  kind: agent
  recorded_at: '2026-11-06T08:18:21.185Z'
id: REQ-ui-inbox-ask-column
type: req
---
The Inbox's open view must offer Ask suggestions about its open items in a side column, each answered on the overview.

## Context

The design session's Inbox mockup adds Ask about these; D-194's box answers them. Modeled by Claude Code for the UI v2 overview slice; the author has not reviewed the wording.

## Source

> **[D-194]** In M0 the Ask box understands a fixed set of commands and answers them from the read API alone: what needs the owner, show an account, sleeve or instrument, why a sync fails, the allocation by a dimension, performance over a period and the value on a date. Each answer is a card with a link to the place it summarizes, and anything else gets the commands it understands as suggestions. It never changes anything, the question never leaves the UI and is never logged (GR-12), and the box says it holds no AI model; in M2 an agent can take it over through the same answer cards (P-57).

docs/DESIGN.md#D-194
