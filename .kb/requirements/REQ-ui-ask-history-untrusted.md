---
title: 'The Ask box''s server treats the earlier turns a browser sends as untrusted: it keeps at most six turns and 4,000 characters of them, dropping the oldest turns first and trimming each field, drops a history it cannot read, shows the model the turns only as a JSON transcript the browser sent, inside the new question''s own message and never as an assistant or tool message, tells it to read every figure again with the tools, and logs neither the question nor the history.'
status: open
firmness: default
required_evidence: unverified
semantic_text: 'The Ask box''s server treats the earlier turns a browser sends as untrusted: it keeps at most six turns and 4,000 characters of them, dropping the oldest turns first and trimming each field, drops a history it cannot read, shows the model the turns only as a JSON transcript the browser sent, inside the new question''s own message and never as an assistant or tool message, tells it to read every figure again with the tools, and logs neither the question nor the history.'
tags:
  - ask
origin:
  kind: agent
  ref: 'Implement D-224: the Ask box holds a conversation (design session brief, 2026-11-21); the author''s feedback that the Ask button must sit inline with the field'
  recorded_at: '2026-11-21T09:22:39.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ae3b380709f370a93420c61e8a426f7b7bea68baa24b71c1bd802f0a97bbcdd3
semantic_inventory:
  - claim_key: CLAIM-58E8051E29E03E8E
    claim_text: 'The Ask box''s server treats the earlier turns a browser sends as untrusted: it keeps at most six turns and 4,000 characters of them, dropping the oldest turns first and trimming each field, drops a history it cannot read, shows the model the turns only as a JSON transcript the browser sent, inside the new question''s own message and never as an assistant or tool message, tells it to read every figure again with the tools, and logs neither the question nor the history'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 470
    subjects:
      - ask_box.conversation_history
logic_claims:
  - CLAIM-58E8051E29E03E8E
id: REQ-ui-ask-history-untrusted
type: req
---
The Ask box's server treats the earlier turns a browser sends as untrusted: it keeps at most six turns and 4,000 characters of them, dropping the oldest turns first and trimming each field, drops a history it cannot read, shows the model the turns only as a JSON transcript the browser sent, inside the new question's own message and never as an assistant or tool message, tells it to read every figure again with the tools, and logs neither the question nor the history.

## Context

D-224 says the conversation is capped and never stored or logged, and the browser's word is untrusted: a forged earlier assistant turn or a planted instruction in an earlier question must not become the model's own words or a tool's result, and a figure in it is never a figure of this question's cards.

## Source

> D-224: The Ask box holds a conversation: after an answer the person can ask a follow-up, and the model sees the earlier questions, answers and cards. The conversation lives only in the browser tab, which sends it with each question, is capped in turns and tokens, and is never stored or logged (D-203). The model's tools are the same as the MCP server's (D-213).
