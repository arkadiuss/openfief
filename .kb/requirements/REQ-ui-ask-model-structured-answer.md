---
title: The Ask box's model ends a question with a structured answer, a JSON object of a kind (answer, clarify or refuse), text, the ids of the cards it read that the answer rests on, the box's commands it offers when it declines and the names to choose from when a name could mean more than one thing, which the loop checks (only cards shown in this question, only commands the box understands, candidates for a clarifying answer), asks for once more when it was plain text and retries once, and a reply that still fails is an error the page shows as the model couldn't answer, never a guess.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box's model ends a question with a structured answer, a JSON object of a kind (answer, clarify or refuse), text, the ids of the cards it read that the answer rests on, the box's commands it offers when it declines and the names to choose from when a name could mean more than one thing, which the loop checks (only cards shown in this question, only commands the box understands, candidates for a clarifying answer), asks for once more when it was plain text and retries once, and a reply that still fails is an error the page shows as the model couldn't answer, never a guess.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7090bbd3411dd54c6a157ea3043041ba8a422abb5729896e479386f00b78f75b
semantic_inventory:
  - claim_key: CLAIM-F42CA4303AE32BCC
    claim_text: The Ask box's model ends a question with a structured answer, a JSON object of a kind (answer, clarify or refuse), text, the ids of the cards it read that the answer rests on, the box's commands it offers when it declines and the names to choose from when a name could mean more than one thing, which the loop checks (only cards shown in this question, only commands the box understands, candidates for a clarifying answer), asks for once more when it was plain text and retries once, and a reply that still fails is an error the page shows as the model couldn't answer, never a guess
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 584
logic_claims:
  - CLAIM-F42CA4303AE32BCC
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:53:05.463Z'
tags:
  - askbench
id: REQ-ui-ask-model-structured-answer
type: req
---
The Ask box's model ends a question with a structured answer, a JSON object of a kind (answer, clarify or refuse), text, the ids of the cards it read that the answer rests on, the box's commands it offers when it declines and the names to choose from when a name could mean more than one thing, which the loop checks (only cards shown in this question, only commands the box understands, candidates for a clarifying answer), asks for once more when it was plain text and retries once, and a reply that still fails is an error the page shows as the model couldn't answer, never a guess.

## Context

A free-text answer cannot be graded exactly: whether it declines, asks which one is meant or names the cards it used had to be read from its words. The design session asked for the final answer as data, validated, so the page shows only what was checked and the bench grades by the schema where it decides (2026-11-14).

## Source

> The AB3 brief from the design session, 2026-11-14: (4) a structured final answer for the Ask box, with strict structured output where the model entry has the flag and JSON mode with one retry where it has not, validated cards and suggestions, the UI drawn from the fields, and the bench scoring by the schema where it decides.