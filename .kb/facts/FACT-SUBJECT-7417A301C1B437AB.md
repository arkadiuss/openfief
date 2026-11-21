---
title: Ask Bench Matcher First
status: active
fact_kind: subject
subject_key: ask_bench.matcher_first
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-7417A301C1B437AB
type: fact
---
The bench step that runs the Ask box's own command matcher before any model, so a question the box answers itself is tagged matcher in the report and counts toward neither safety nor quality. It does not cover questions that go to a model.