---
title: The Ask box bench treats a model that no call reached as a run that could not run, exiting 2 and saying so per model, scores a partial outage on the questions that were answered, and lists each failed call with its HTTP status and a short fixed label, never the provider's reply or the key, so that a failed call is never counted as a model failure.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box bench treats a model that no call reached as a run that could not run, exiting 2 and saying so per model, scores a partial outage on the questions that were answered, and lists each failed call with its HTTP status and a short fixed label, never the provider's reply or the key, so that a failed call is never counted as a model failure.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: aa282f0038729e219af7607d782593c6eb82c4fd28fd727c0a42422e686aa1ca
semantic_inventory:
  - claim_key: CLAIM-14EEBFF79E9CE907
    claim_text: The Ask box bench treats a model that no call reached as a run that could not run, exiting 2 and saying so per model, scores a partial outage on the questions that were answered, and lists each failed call with its HTTP status and a short fixed label, never the provider's reply or the key, so that a failed call is never counted as a model failure
    role: rationale
    status: nonlogical
    span:
      start: 0
      end: 348
origin:
  kind: agent
  ref: Brief to the Ask bench status session (a run no model answered must not look like a model result), 2026-11-14
  recorded_at: '2026-11-14T13:26:24.643Z'
tags:
  - askbench
id: REQ-askbench-errors-not-model-failures
type: req
---
The Ask box bench treats a model that no call reached as a run that could not run, exiting 2 and saying so per model, scores a partial outage on the questions that were answered, and lists each failed call with its HTTP status and a short fixed label, never the provider's reply or the key, so that a failed call is never counted as a model failure.

## Context

When every call failed (a refused key, no credits, no network) the bench exited 1, as if a safety case had failed, and its report listed model failures where the model had never been asked. The status of each failed call says which kind of outage it was.

## Source

> A run where NOTHING reached a model must not look like a model result. A partial outage (some calls fail) stays scored, but the report must list those calls as errors, never as model failures. Record the HTTP status of each failed model call: status code and a short fixed label, NEVER the provider's response body, never the key.

The brief of this change, 2026-11-14.
