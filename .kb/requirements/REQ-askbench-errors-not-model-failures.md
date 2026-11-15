---
title: The Ask box bench treats a model that no call reached as a run that could not run, saying so per model and exiting 2 (or 5 when a refused key or no credits voided it), scores a partial outage on the questions that were answered, and lists each failed call with its HTTP status and a short fixed label, never the provider's reply or the key, so that a failed call is never counted as a model failure.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box bench treats a model that no call reached as a run that could not run, saying so per model and exiting 2 (or 5 when a refused key or no credits voided it), scores a partial outage on the questions that were answered, and lists each failed call with its HTTP status and a short fixed label, never the provider's reply or the key, so that a failed call is never counted as a model failure.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4f180a0f13d202157a9e6e450c25d72d50802d5796e66bbe137844221bf07f24
semantic_inventory:
  - claim_key: CLAIM-6F28F52271C5F13F
    claim_text: The Ask box bench treats a model that no call reached as a run that could not run, saying so per model and exiting 2 (or 5 when a refused key or no credits voided it), scores a partial outage on the questions that were answered, and lists each failed call with its HTTP status and a short fixed label, never the provider's reply or the key, so that a failed call is never counted as a model failure
    role: rationale
    status: nonlogical
    span:
      start: 0
      end: 398
origin:
  kind: agent
  ref: 'Brief of the Ask bench 4xx and per-question reliability change: the design delegate''s decisions on D-207 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T20:00:00.000Z'
tags:
  - askbench
id: REQ-askbench-errors-not-model-failures
type: req
---
The Ask box bench treats a model that no call reached as a run that could not run, saying so per model and exiting 2 (or 5 when a refused key or no credits voided it), scores a partial outage on the questions that were answered, and lists each failed call with its HTTP status and a short fixed label, never the provider's reply or the key, so that a failed call is never counted as a model failure.

## Context

When every call failed (a refused key, no credits, no network) the bench exited 1, as if a safety case had failed, and its report listed model failures where the model had never been asked. The status of each failed call says which kind of outage it was.

## Source

> A run where NOTHING reached a model must not look like a model result. A partial outage (some calls fail) stays scored, but the report must list those calls as errors, never as model failures. Record the HTTP status of each failed model call: status code and a short fixed label, NEVER the provider's response body, never the key.

The brief of this change, 2026-11-14.
