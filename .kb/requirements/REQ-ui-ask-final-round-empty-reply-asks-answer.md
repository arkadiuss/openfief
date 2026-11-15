---
title: When the model's reply before the Ask box's final round has no text (it only called tools, or stopped after a tool's result with nothing or only blanks), the final round sends no empty assistant message and asks for the answer itself in the schema, using only the cards read, instead of asking to put a previous reply into the schema; a reply with text keeps the instruction to put it into the schema without changing its meaning, and the retry after an invalid reply keeps fixing only the form, for every model and with no hint for any one question.
status: open
tags:
  - askbench
origin:
  kind: agent
  ref: Design session approval of the empty-reply fix after bench traces (D-205, D-217), 2026-11-15
  recorded_at: '2026-11-15T13:00:11.533Z'
firmness: default
required_evidence: unverified
id: REQ-ui-ask-final-round-empty-reply-asks-answer
type: req
semantic_text: When the model's reply before the Ask box's final round has no text (it only called tools, or stopped after a tool's result with nothing or only blanks), the final round sends no empty assistant message and asks for the answer itself in the schema, using only the cards read, instead of asking to put a previous reply into the schema; a reply with text keeps the instruction to put it into the schema without changing its meaning, and the retry after an invalid reply keeps fixing only the form, for every model and with no hint for any one question.
logic_claims:
  - CLAIM-28B0735D3ACC6F2F
  - CLAIM-F3D732D0D9547651
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 701cc18e09a90cdf8cf02a2172da5c8a62af5437a9006b38459670fe19c76460
semantic_inventory:
  - claim_key: CLAIM-28B0735D3ACC6F2F
    claim_text: When the model's reply before the Ask box's final round has no text (it only called tools, or stopped after a tool's result with nothing or only blanks), the final round sends no empty assistant message and asks for the answer itself in the schema, using only the cards read, instead of asking to put a previous reply into the schema
    role: condition
    status: ontology_gap
    span:
      start: 0
      end: 333
  - claim_key: CLAIM-F3D732D0D9547651
    claim_text: a reply with text keeps the instruction to put it into the schema without changing its meaning, and the retry after an invalid reply keeps fixing only the form, for every model and with no hint for any one question
    role: descriptive
    status: missing
    span:
      start: 335
      end: 549
---
When the model's reply before the Ask box's final round has no text (it only called tools, or stopped after a tool's result with nothing or only blanks), the final round sends no empty assistant message and asks for the answer itself in the schema, using only the cards read, instead of asking to put a previous reply into the schema; a reply with text keeps the instruction to put it into the schema without changing its meaning, and the retry after an invalid reply keeps fixing only the form, for every model and with no hint for any one question.

## Context

Bench traces showed round 1 calling a tool, round 2 returning an empty reply (finish_reason stop, no text, no tool calls), and the final round then asking to put the previous reply into the schema. There was none, and the model refused ("I'm sorry, I can't modify my previous reply."). The restating instruction (D-205, D-217) stays for a reply that has text.

## Source

> Design session approval, 2026-11-15: when the last reply is empty the final round says to answer the question in the schema, using only the cards above.