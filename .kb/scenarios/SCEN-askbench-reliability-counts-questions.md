---
title: Reliability counts a question once however many calls failed, and re-scoring counts it again from the stored failures
status: active
origin:
  kind: agent
  ref: 'Brief of the Ask bench 4xx and per-question reliability change: the design delegate''s decisions on D-207 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T20:00:00.000Z'
id: SCEN-askbench-reliability-counts-questions
type: scenario
---
Given a host that answers the final round of every question with a reply that is no chat completion, then each question's call and its retry both fail, the question counts once as failed, and reliability reads five of five questions and ten failed calls, not ten of twenty; given a call that its retry mended, then the question is not failed; given a question the matcher answered, one the budget stopped or a void run, then it is not counted; and given a report written before scorer version 3, then re-scoring counts reliability again per question from the stored call failures, derives the faults it can from their status and label, says how many 4xx it could not decide, and leaves the configuration hash as it was.