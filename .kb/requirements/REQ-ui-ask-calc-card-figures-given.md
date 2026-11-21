---
title: The calc tool must read no data of the owner and answer with a card that shows the expression with its result, says that the result is computed from the figures given, and carries no data tier, so that a route cleared for T0 only is offered it and its result passes, and the tool must take only an expression, in the toolbox of the Ask box and of the MCP server alike.
status: open
firmness: default
origin:
  kind: agent
  ref: D-228 as amended; the coordinator's correction that calc adds no data tier, 2026-11-21
  recorded_at: '2026-11-21T16:00:00.000Z'
tags:
  - ui
  - ask
  - calc
  - d-228
text_ref: docs/DESIGN.md#D-228
semantic_text: The calc tool must read no data of the owner and answer with a card that shows the expression with its result, says that the result is computed from the figures given, and carries no data tier, so that a route cleared for T0 only is offered it and its result passes, and the tool must take only an expression, in the toolbox of the Ask box and of the MCP server alike.
required_evidence: unverified
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e9c09a268772d86fcc37a6b5c392265a878f916243316cfe06647335d86d4c9c
semantic_inventory:
  - claim_key: CLAIM-2042B607EF1F5F59
    claim_text: The calc tool must read no data of the owner and answer with a card that shows the expression with its result, says that the result is computed from the figures given, and carries no data tier, so that a route cleared for T0 only is offered it and its result passes, and the tool must take only an expression, in the toolbox of the Ask box and of the MCP server alike
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 367
    subjects:
      - ask_box.calc_card
logic_claims:
  - CLAIM-2042B607EF1F5F59
rationale: 'The figures of a calc call come from the model and are already in the conversation, cards the person may see or text they wrote, so the card adds no data and no tier: tagged as the box''s own text (D-226) it never makes a T0-only route lose calc, which would be the opposite of D-226''s ruling that such a route answers everything but prices (D-228, D-206, D-213).'
id: REQ-ui-ask-calc-card-figures-given
type: req
---
The calc tool must read no data of the owner and answer with a card that shows the expression with its result, says that the result is computed from the figures given, and carries no data tier, so that a route cleared for T0 only is offered it and its result passes, and the tool must take only an expression, in the toolbox of the Ask box and of the MCP server alike.

## Context

The figures of a calc call come from the model and are already in the conversation, cards the person may see or text they wrote, so the card adds no data and no tier: tagged as the box's own text (D-226) it never makes a T0-only route lose calc, which would be the opposite of D-226's ruling that such a route answers everything but prices (D-228, D-206, D-213).

## Source

> The Ask box and the MCP server also offer a tool that does exact decimal arithmetic on figures already shown.

docs/DESIGN.md#D-228
