---
title: The Ask box must understand what needs me, show an account, sleeve or instrument by name, ticker or ISIN, why a sync fails, allocation by a dimension, performance over a period and value on a date, with suggestion chips drawn from the current state, and focus on the slash key.
status: open
text_ref: docs/DESIGN.md#D-194
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-194
  - lane:strict
semantic_text: The Ask box must understand what needs me, show an account, sleeve or instrument by name, ticker or ISIN, why a sync fails, allocation by a dimension, performance over a period and value on a date, with suggestion chips drawn from the current state, and focus on the slash key.
logic_claims:
  - CLAIM-D959F9468739FF12
semantic_clauses:
  - The Ask box must understand what needs me, show an account, sleeve or instrument by name, ticker or ISIN, why a sync fails, allocation by a dimension, performance over a period and value on a date, with suggestion chips drawn from the current state, and focus on the slash key.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 13299936efd9f27942298c675d49e3b219e64ea92baa1377a17d2b2dee35230a
semantic_inventory:
  - claim_key: CLAIM-D959F9468739FF12
    claim_text: The Ask box must understand what needs me, show an account, sleeve or instrument by name, ticker or ISIN, why a sync fails, allocation by a dimension, performance over a period and value on a date, with suggestion chips drawn from the current state, and focus on the slash key
    role: normative
    status: modeled
    span:
      start: 0
      end: 276
origin:
  kind: agent
  recorded_at: '2026-11-06T08:18:13.472Z'
id: REQ-ui-ask-commands
type: req
---
The Ask box must understand what needs me, show an account, sleeve or instrument by name, ticker or ISIN, why a sync fails, allocation by a dimension, performance over a period and value on a date, with suggestion chips drawn from the current state, and focus on the slash key.

## Context

Refines D-194's commands into the phrasings parseAsk accepts. Modeled by Claude Code for the UI v2 overview slice; the author has not reviewed the wording.

## Source

> **[D-194]** In M0 the Ask box understands a fixed set of commands and answers them from the read API alone: what needs the owner, show an account, sleeve or instrument, why a sync fails, the allocation by a dimension, performance over a period and the value on a date. Each answer is a card with a link to the place it summarizes, and anything else gets the commands it understands as suggestions. It never changes anything, the question never leaves the UI and is never logged (GR-12), and the box says it holds no AI model; in M2 an agent can take it over through the same answer cards (P-57).

docs/DESIGN.md#D-194
