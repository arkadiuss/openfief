---
title: 'The Ask box holds a conversation: each answer is added to a thread under the field, a follow-up carries the last six earlier turns, as the question, the answer''s text and compact summaries of its cards, in the body of the POST that asks it and never in a URL, the turns exist only in the page''s memory, never in localStorage, sessionStorage or a cookie, New conversation clears the thread and with it what the next question carries, a command is still answered by the box itself inside a conversation and joins it, and each model answer keeps its own single foot note.'
status: open
firmness: default
required_evidence: unverified
semantic_text: 'The Ask box holds a conversation: each answer is added to a thread under the field, a follow-up carries the last six earlier turns, as the question, the answer''s text and compact summaries of its cards, in the body of the POST that asks it and never in a URL, the turns exist only in the page''s memory, never in localStorage, sessionStorage or a cookie, New conversation clears the thread and with it what the next question carries, a command is still answered by the box itself inside a conversation and joins it, and each model answer keeps its own single foot note.'
tags:
  - ask
origin:
  kind: agent
  ref: 'Implement D-224: the Ask box holds a conversation (design session brief, 2026-11-21); the author''s feedback that the Ask button must sit inline with the field'
  recorded_at: '2026-11-21T09:22:39.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 21181eb8dab1083bcead52ba52440b31fa569fae948c428807373f69e5d25ec8
semantic_inventory:
  - claim_key: CLAIM-EA8E29477A9D4D37
    claim_text: 'The Ask box holds a conversation: each answer is added to a thread under the field, a follow-up carries the last six earlier turns, as the question, the answer''s text and compact summaries of its cards, in the body of the POST that asks it and never in a URL, the turns exist only in the page''s memory, never in localStorage, sessionStorage or a cookie, New conversation clears the thread and with it what the next question carries, a command is still answered by the box itself inside a conversation and joins it, and each model answer keeps its own single foot note'
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 567
    subjects:
      - ask_box.conversation_thread
logic_claims:
  - CLAIM-EA8E29477A9D4D37
id: REQ-ui-ask-conversation
type: req
---
The Ask box holds a conversation: each answer is added to a thread under the field, a follow-up carries the last six earlier turns, as the question, the answer's text and compact summaries of its cards, in the body of the POST that asks it and never in a URL, the turns exist only in the page's memory, never in localStorage, sessionStorage or a cookie, New conversation clears the thread and with it what the next question carries, a command is still answered by the box itself inside a conversation and joins it, and each model answer keeps its own single foot note.

## Context

D-224 makes the Ask box a conversation kept only in the browser tab. The history is read from the thread the page shows (data-turn on each answer) when a question is asked, sent as a POST so that no question or earlier turn is ever in a URL a proxy might log, and cleared by New conversation. The Turnstile check is unchanged and reads its token from the POST body as from the old query.

## Source

> D-224: The Ask box holds a conversation: after an answer the person can ask a follow-up, and the model sees the earlier questions, answers and cards. The conversation lives only in the browser tab, which sends it with each question, is capped in turns and tokens, and is never stored or logged (D-203). The model's tools are the same as the MCP server's (D-213).
