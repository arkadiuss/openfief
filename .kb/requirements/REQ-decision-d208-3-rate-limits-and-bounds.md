---
title: The demo's model spend has per-session and per-IP rate limits of 10 questions an hour and 30 a day, at most 3 model calls at once, 800 output tokens, 5 tool rounds and a timeout per question, and the maintainer only on its weekly schedule.
status: open
firmness: firm
semantic_text: The demo's model spend has per-session and per-IP rate limits of 10 questions an hour and 30 a day, at most 3 model calls at once, 800 output tokens, 5 tool rounds and a timeout per question, and the maintainer only on its weekly schedule.
origin:
  kind: import
  ref: design-decisions:D-208#3
  recorded_at: '2026-11-14T20:37:38.985Z'
text_ref: design-decisions:D-208#3
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 13d13dac81c4f25b7a1ec1f3bccbe85517e430e0b67170b0505f16549256ff61
semantic_inventory:
  - claim_key: CLAIM-71460C92F8D4FD79
    claim_text: The demo's model spend has per-session and per-IP rate limits of 10 questions an hour and 30 a day, at most 3 model calls at once, 800 output tokens, 5 tool rounds and a timeout per question, and the maintainer only on its weekly schedule
    role: descriptive
    status: missing
    span:
      start: 0
      end: 238
logic_claims:
  - CLAIM-71460C92F8D4FD79
id: REQ-decision-d208-3-rate-limits-and-bounds
type: req
---
The demo's model spend has per-session and per-IP rate limits of 10 questions an hour and 30 a day, at most 3 model calls at once, 800 output tokens, 5 tool rounds and a timeout per question, and the maintainer only on its weekly schedule.

## Context

Author decision D-208 (the demo's spend guards), imported from docs/DESIGN.md; it was not imported before. REQ-models-ask-question-limits, REQ-models-ask-concurrent-calls and REQ-ui-ask-model-loop-limits are the behavior; the demo's values are set in deploy/demo/model.

## Source

> docs/DESIGN.md, D-208#3.