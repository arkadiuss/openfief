---
title: Ask Bench Transport Retries
status: active
fact_kind: subject
subject_key: ask_bench.transport_retries
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-70BC7E662F9C932B
type: fact
---
The bench's retry rules for a call that failed with no answer, a timeout, a 5xx or a bad reply, and for a 429, and the reliability it reports from them; a 4xx or an answer is never retried. It does not cover how a 4xx is blamed.