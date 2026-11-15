---
title: 'The Ask bench''s report traces each model round of a run: the round number, the roles, count, length and short hash of the messages sent, the prompt, completion and cached tokens the host reported, the finish reason, the names of the tools asked for and the reply''s plain text cut to 500 characters, and never the text of what was sent.'
status: open
firmness: default
required_evidence: unverified
tags:
  - askbench
origin:
  kind: agent
  ref: 'Live bench finding on R1 (gpt-oss-120b on deepinfra: two model calls, 1694 input tokens, a greeting), 2026-11-14'
  recorded_at: '2026-11-14T23:19:42.286Z'
semantic_text: 'The Ask bench''s report.json holds, under each run''s rounds, one entry for each call the run made to the model: the round number, the roles and count of the messages sent with each one''s length in characters and a short hash of its content, the prompt, completion and cached tokens the host reported, the reply''s finish reason, the names of the tools it asked for and its plain text cut to 500 characters; it holds no text of what was sent.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ac240f90a3d9577f09c3a51d69f346cc3be9e16df8cb54a3608dade475e05fed
semantic_inventory:
  - claim_key: CLAIM-805D2087C3E508B6
    claim_text: 'The Ask bench''s report.json holds, under each run''s rounds, one entry for each call the run made to the model: the round number, the roles and count of the messages sent with each one''s length in characters and a short hash of its content, the prompt, completion and cached tokens the host reported, the reply''s finish reason, the names of the tools it asked for and its plain text cut to 500 characters'
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 403
  - claim_key: CLAIM-5EE1A3DFE2AB46EB
    claim_text: it holds no text of what was sent
    role: descriptive
    status: ontology_gap
    span:
      start: 405
      end: 438
logic_claims:
  - CLAIM-805D2087C3E508B6
  - CLAIM-5EE1A3DFE2AB46EB
id: REQ-askbench-traces-each-model-round
type: req
---
The Ask bench's report.json holds, under each run's rounds, one entry for each call the run made to the model: the round number, the roles and count of the messages sent with each one's length in characters and a short hash of its content, the prompt, completion and cached tokens the host reported, the reply's finish reason, the names of the tools it asked for and its plain text cut to 500 characters; it holds no text of what was sent.

## Context

On gpt-oss-120b pinned to deepinfra/bf16, case R1 failed 3 of 5 runs after two model calls totalling 1694 input tokens, while a passing single call took 1018. The first call reported under 850 input tokens: either the host reports only uncached tokens or part of the prompt did not arrive. The trace tells the two apart. The non-public part of the prompt must stay out of report.json, so messages are traced by role, length and hash only.

## Source

Asked in the agent brief: "record, per run of a case, each model round ... Do not record full prompt text (the non-public prompt part must stay out of report.json)".