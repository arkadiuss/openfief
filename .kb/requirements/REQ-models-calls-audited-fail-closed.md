---
title: Each model call and each refusal of one enters the audit log naming the actor, the role and the model by name and the outcome, never a question, an answer, an amount or a token count, and the router refuses a call whose entry cannot be written.
status: open
firmness: default
origin:
  kind: agent
  ref: the M1 brief for the model slice (D-205, D-206, D-211, D-212)
  recorded_at: '2026-11-13T11:28:55.570Z'
semantic_text: Each model call and each refusal of one enters the audit log naming the actor, the role and the model by name and the outcome, never a question, an answer, an amount or a token count, and the router refuses a call whose entry cannot be written.
semantic_source_field: semantic_text
semantic_source_hash: b3cb954d57835fe5a3e8b8ae6fc846e08f3501ce9690123786890cd7af04edd3
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-06807EA982FF01CC
    claim_text: Each model call and each refusal of one enters the audit log naming the actor, the role and the model by name and the outcome, never a question, an answer, an amount or a token count, and the router refuses a call whose entry cannot be written
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 243
logic_claims:
  - CLAIM-06807EA982FF01CC
semantic_clauses:
  - Each model call and each refusal of one enters the audit log naming the actor, the role and the model by name and the outcome, never a question, an answer, an amount or a token count, and the router refuses a call whose entry cannot be written
required_evidence: unverified
id: REQ-models-calls-audited-fail-closed
type: req
---
Each model call and each refusal of one enters the audit log naming the actor, the role and the model by name and the outcome, never a question, an answer, an amount or a token count, and the router refuses a call whose entry cannot be written.

## Context

D-211 puts model calls in the append-only audit log, whose entries never carry an amount, a secret or a question's text (GR-12); the M1 brief adds that the router fails closed.

## Source

> model calls are audit events; fail closed: refuse the call if its entry can't be written

Source: the M1 brief for the model slice (D-205, D-206, D-211, D-212)