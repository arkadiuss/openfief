---
title: The calc tool must compute an expression of decimal numbers with addition, subtraction, multiplication, division, signs, parentheses and percent signs exactly in decimals, so that 0.1 plus 0.2 is 0.3, rounding only a division that does not end within 20 decimal places, half-even and once, and saying that it did.
status: open
firmness: default
origin:
  kind: agent
  ref: 'D-228 as amended (the Ask box and the MCP server offer a tool for exact decimal arithmetic): the design session''s brief, 2026-11-21'
  recorded_at: '2026-11-21T14:10:00.000Z'
tags:
  - ui
  - ask
  - calc
  - d-228
text_ref: docs/DESIGN.md#D-228
semantic_text: The calc tool must compute an expression of decimal numbers with addition, subtraction, multiplication, division, signs, parentheses and percent signs exactly in decimals, so that 0.1 plus 0.2 is 0.3, rounding only a division that does not end within 20 decimal places, half-even and once, and saying that it did.
required_evidence: unverified
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 64c8808c8d58048340d9a48a67b2773f36fa6dafc293576fe414194a27481955
semantic_inventory:
  - claim_key: CLAIM-6343F7088972C86E
    claim_text: The calc tool must compute an expression of decimal numbers with addition, subtraction, multiplication, division, signs, parentheses and percent signs exactly in decimals, so that 0.1 plus 0.2 is 0.3, rounding only a division that does not end within 20 decimal places, half-even and once, and saying that it did
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 312
    subjects:
      - ask_box.calc_exactness
logic_claims:
  - CLAIM-6343F7088972C86E
rationale: Models do arithmetic with floating point or by guessing, and a sum of money must be exact. D-228 gives the Ask box and the MCP server one tool that does exact decimal arithmetic on figures already shown, so a model can work out, for example, the amounts to move toward target weights.
id: REQ-ui-ask-calc-exact
type: req
---
The calc tool must compute an expression of decimal numbers with addition, subtraction, multiplication, division, signs, parentheses and percent signs exactly in decimals, so that 0.1 plus 0.2 is 0.3, rounding only a division that does not end within 20 decimal places, half-even and once, and saying that it did.

## Context

Models do arithmetic with floating point or by guessing, and a sum of money must be exact. D-228 gives the Ask box and the MCP server one tool that does exact decimal arithmetic on figures already shown, so a model can work out, for example, the amounts to move toward target weights.

## Source

> The Ask box and the MCP server also offer a tool that does exact decimal arithmetic on figures already shown.

docs/DESIGN.md#D-228
