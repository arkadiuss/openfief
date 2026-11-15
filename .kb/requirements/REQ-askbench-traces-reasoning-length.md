---
title: The Ask box bench's trace of each model round records the length in characters of the reasoning the host returned with the reply (message.reasoning, else message.reasoning_content), never its text, and the chat client never uses reasoning as an answer.
status: open
tags:
  - askbench
origin:
  kind: agent
  ref: Design session approval of the empty-reply fix after bench traces (D-205, D-217), 2026-11-15
  recorded_at: '2026-11-15T13:00:12.099Z'
firmness: default
required_evidence: unverified
id: REQ-askbench-traces-reasoning-length
type: req
semantic_text: The Ask box bench's trace of each model round records the length in characters of the reasoning the host returned with the reply (message.reasoning, else message.reasoning_content), never its text, and the chat client never uses reasoning as an answer.
logic_claims:
  - CLAIM-7A88A2E5F2C691FD
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 64d0227ad73f6f138ced87206cba80d3f812aaa0eda77098780f8c7d050cbbd1
semantic_inventory:
  - claim_key: CLAIM-7A88A2E5F2C691FD
    claim_text: The Ask box bench's trace of each model round records the length in characters of the reasoning the host returned with the reply (message.reasoning, else message.reasoning_content), never its text, and the chat client never uses reasoning as an answer
    role: descriptive
    status: missing
    span:
      start: 0
      end: 251
---
The Ask box bench's trace of each model round records the length in characters of the reasoning the host returned with the reply (message.reasoning, else message.reasoning_content), never its text, and the chat client never uses reasoning as an answer.

## Context

A model that returns an empty reply may have written its thoughts in a reasoning field. The trace says whether it did, by length, so an empty round can be told from a silent one without the report holding text that is never an answer.

## Source

> Design session instruction, 2026-11-15: record per round whether the response's reasoning field was non-empty, as a length only, never its text.