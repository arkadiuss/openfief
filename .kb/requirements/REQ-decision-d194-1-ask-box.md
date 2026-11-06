---
title: The Ask box must answer a fixed set of commands from the read API alone, each as a card with a link to the place it summarizes, give anything else the commands it understands, change nothing, never log or send the question anywhere, and say it holds no AI model.
status: open
text_ref: design-decisions:D-194#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-194-1
  - lane:strict
semantic_text: The Ask box must answer a fixed set of commands from the read API alone, each as a card with a link to the place it summarizes, give anything else the commands it understands, change nothing, never log or send the question anywhere, and say it holds no AI model.
logic_claims:
  - CLAIM-873D0BA3262A497E
semantic_clauses:
  - The Ask box must answer a fixed set of commands from the read API alone, each as a card with a link to the place it summarizes, give anything else the commands it understands, change nothing, never log or send the question anywhere, and say it holds no AI model.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e18a941dbb20b5afe20fe339a2dffcbe28c940ae7dcf1bdaf1199780aeb428dc
semantic_inventory:
  - claim_key: CLAIM-873D0BA3262A497E
    claim_text: The Ask box must answer a fixed set of commands from the read API alone, each as a card with a link to the place it summarizes, give anything else the commands it understands, change nothing, never log or send the question anywhere, and say it holds no AI model
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-194#1
  recorded_at: '2026-11-06T08:18:02.153Z'
id: REQ-decision-d194-1-ask-box
type: req
---
The Ask box must answer a fixed set of commands from the read API alone, each as a card with a link to the place it summarizes, give anything else the commands it understands, change nothing, never log or send the question anywhere, and say it holds no AI model.

## Context

Author decision D-194 (UI v2 design session).

## Source

> **[D-194]** In M0 the Ask box understands a fixed set of commands and answers them from the read API alone: what needs the owner, show an account, sleeve or instrument, why a sync fails, the allocation by a dimension, performance over a period and the value on a date. Each answer is a card with a link to the place it summarizes, and anything else gets the commands it understands as suggestions. It never changes anything, the question never leaves the UI and is never logged (GR-12), and the box says it holds no AI model; in M2 an agent can take it over through the same answer cards (P-57).

docs/DESIGN.md#D-194
