---
title: 'A model call is two audit entries linked by a call ID: model_call with the outcome sent, written before the call is sent, and model_call_result after it with ok, failed with an error class, or refused; a call the router refuses before choosing a model has only the result entry; both name the role, the provider instance and model by name and never a question, an answer or a token.'
status: open
firmness: default
origin:
  kind: agent
  ref: design-session answers for the M1 merge and follow-ups, 2026-11-13 (D-211, D-205)
  recorded_at: '2026-11-13T14:45:43.946Z'
semantic_text: 'A model call is two audit entries linked by a call ID: model_call with the outcome sent, written before the call is sent, and model_call_result after it with ok, failed with an error class, or refused; a call the router refuses before choosing a model has only the result entry; both name the role, the provider instance and model by name and never a question, an answer or a token.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c4ed5479534fe8ced4c5c2d74938309a4f3e1406c6021320aa9449e3648cbe13
semantic_inventory:
  - claim_key: CLAIM-2D5316F15AFCC6D8
    claim_text: 'A model call is two audit entries linked by a call ID: model_call with the outcome sent, written before the call is sent'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 120
  - claim_key: CLAIM-9429547524C35B95
    claim_text: model_call_result after it with ok, failed with an error class, or refused
    role: normative
    status: ontology_gap
    span:
      start: 126
      end: 200
  - claim_key: CLAIM-5FDF3660B4651869
    claim_text: a call the router refuses before choosing a model has only the result entry
    role: normative
    status: ontology_gap
    span:
      start: 202
      end: 277
  - claim_key: CLAIM-F05F654C2F819F0D
    claim_text: both name the role, the provider instance and model by name and never a question, an answer or a token
    role: descriptive
    status: missing
    span:
      start: 279
      end: 381
logic_claims:
  - CLAIM-2D5316F15AFCC6D8
  - CLAIM-9429547524C35B95
  - CLAIM-5FDF3660B4651869
  - CLAIM-F05F654C2F819F0D
id: REQ-models-calls-two-entries-linked
type: req
---
A model call is two audit entries linked by a call ID: model_call with the outcome sent, written before the call is sent, and model_call_result after it with ok, failed with an error class, or refused; a call the router refuses before choosing a model has only the result entry; both name the role, the provider instance and model by name and never a question, an answer or a token.

## Context

D-211 puts model calls in the audit log and the router fails closed: a call whose entry can't be written is not made. The design session that answered the M1 follow-ups split the entry in two, so a call that never came back is visible as a sent entry without a result. Usage and cost stay with the budget (D-208).

## Source

> Audit: two entries linked by a call ID: `model_call` outcome `sent` before the call (fail closed: not written -> no call); `model_call_result` after with ok | error class | refused. Router denial by tier/guards: only the result entry (refused). Role, provider instance, model only; never question/answer/token text. Usage/cost stay with D-208.

Source: design-session answers for the M1 merge and follow-ups, 2026-11-13