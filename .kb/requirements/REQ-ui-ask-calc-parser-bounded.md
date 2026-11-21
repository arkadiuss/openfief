---
title: The calc tool must read an expression only in a fixed grammar of numbers, the four operations, percent signs, signs and parentheses and evaluate nothing else, and must refuse, with a reason that names no figure of the expression, one that is empty, longer than 200 characters, nested deeper than 20 levels, holds a number of more than 30 digits, makes a result of more than 60 digits, or divides by zero.
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
semantic_text: The calc tool must read an expression only in a fixed grammar of numbers, the four operations, percent signs, signs and parentheses and evaluate nothing else, and must refuse, with a reason that names no figure of the expression, one that is empty, longer than 200 characters, nested deeper than 20 levels, holds a number of more than 30 digits, makes a result of more than 60 digits, or divides by zero.
required_evidence: unverified
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e99fe5dea014dc0939cb4c350f0c5e67eea2d9256ba78dbe0e16882e33429b46
semantic_inventory:
  - claim_key: CLAIM-94E0DED444A63B67
    claim_text: The calc tool must read an expression only in a fixed grammar of numbers, the four operations, percent signs, signs and parentheses and evaluate nothing else, and must refuse, with a reason that names no figure of the expression, one that is empty, longer than 200 characters, nested deeper than 20 levels, holds a number of more than 30 digits, makes a result of more than 60 digits, or divides by zero
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 403
    subjects:
      - ask_box.calc_parser
logic_claims:
  - CLAIM-94E0DED444A63B67
rationale: The expression comes from a model, which a stored name or a forged turn may have instructed (D-203, R-6). A parser that evaluates only a closed grammar and is bounded cannot run code, run long or grow a number without end, and an error that quotes no figure keeps the owner's amounts out of logs and results (GR-12).
id: REQ-ui-ask-calc-parser-bounded
type: req
---
The calc tool must read an expression only in a fixed grammar of numbers, the four operations, percent signs, signs and parentheses and evaluate nothing else, and must refuse, with a reason that names no figure of the expression, one that is empty, longer than 200 characters, nested deeper than 20 levels, holds a number of more than 30 digits, makes a result of more than 60 digits, or divides by zero.

## Context

The expression comes from a model, which a stored name or a forged turn may have instructed (D-203, R-6). A parser that evaluates only a closed grammar and is bounded cannot run code, run long or grow a number without end, and an error that quotes no figure keeps the owner's amounts out of logs and results (GR-12).

## Source

> The Ask box and the MCP server also offer a tool that does exact decimal arithmetic on figures already shown.

docs/DESIGN.md#D-228
