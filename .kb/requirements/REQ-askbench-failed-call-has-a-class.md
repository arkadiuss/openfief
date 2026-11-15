---
title: The Ask bench's report gives every failed call to the model an error class from a fixed set, in each failed call's entry in report.json and in the failed-calls line of summary.md and stdout, never the provider's reply, so that no call fails as a bare failed; a failed call is the host's fault, counted in the reliability, for no answer, a timeout, a 5xx, a 429, a degenerate generation or a bad reply to a request the bench checked is well-formed, and the bench's fault, which makes the run void and keeps it out of the history, for a request that breaks the chat API's rules or a local error.
status: open
firmness: default
required_evidence: unverified
tags:
  - askbench
origin:
  kind: agent
  ref: 'Brief on run 16-1: three calls failed as "failed" with no cause in report.json (D-207, D-217), 2026-11-15'
  recorded_at: '2026-11-15T18:32:24.665Z'
semantic_text: The Ask bench's report gives every failed call to the model an error class from a fixed set, in each failed call's entry in report.json and in the failed-calls line of summary.md and stdout, never the provider's reply, so that no call fails as a bare failed; a failed call is the host's fault, counted in the reliability, for no answer, a timeout, a 5xx, a 429, a degenerate generation or a bad reply to a request the bench checked is well-formed, and the bench's fault, which makes the run void and keeps it out of the history, for a request that breaks the chat API's rules or a local error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7b5d3d3b3ecaca56579638e59fdda364c7c6190c8015aa28ba34ac39646182da
semantic_inventory:
  - claim_key: CLAIM-731D8E3CBF9975A9
    claim_text: The Ask bench's report gives every failed call to the model an error class from a fixed set, in each failed call's entry in report.json and in the failed-calls line of summary.md and stdout, never the provider's reply, so that no call fails as a bare failed
    role: rationale
    status: nonlogical
    span:
      start: 0
      end: 257
  - claim_key: CLAIM-74250365541CA78A
    claim_text: a failed call is the host's fault, counted in the reliability, for no answer, a timeout, a 5xx, a 429, a degenerate generation or a bad reply to a request the bench checked is well-formed, and the bench's fault, which makes the run void and keeps it out of the history, for a request that breaks the chat API's rules or a local error
    role: descriptive
    status: ontology_gap
    span:
      start: 259
      end: 592
logic_claims:
  - CLAIM-74250365541CA78A
id: REQ-askbench-failed-call-has-a-class
type: req
---
The Ask bench's report gives every failed call to the model an error class from a fixed set, in each failed call's entry in report.json and in the failed-calls line of summary.md and stdout, never the provider's reply, so that no call fails as a bare failed; a failed call is the host's fault, counted in the reliability, for no answer, a timeout, a 5xx, a 429, a degenerate generation or a bad reply to a request the bench checked is well-formed, and the bench's fault, which makes the run void and keeps it out of the history, for a request that breaks the chat API's rules or a local error.

## Context

In run 16-1 (gpt-oss-120b on deepinfra/bf16) three calls failed with the label "failed": no HTTP status, no timeout and not a chat-client call error. The class was computed but never written to report.json, so the cause was invisible, and those calls were neither transport failures (D-207) nor scored runs. Nothing may leave a run silently unscored: each failed call is in one of three buckets (the host's, the bench's, the model's).

## Source

The brief of this change and the design session's approval, 2026-11-15: "every failed call gets a fixed-vocabulary class and lands in exactly one bucket: host fault counts toward D-207 reliability; our fault makes the run VOID, outside D-221's streak, re-run after the fix".