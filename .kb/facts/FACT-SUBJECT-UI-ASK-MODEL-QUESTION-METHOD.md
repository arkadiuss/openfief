---
title: UI Ask Model Question Method
status: active
fact_kind: subject
subject_key: ui.ask_model_question_method
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #9: a model question must need POST so a same-site GET can''t spend model budget (the coordinator''s follow-up, 2026-11-21)'
  recorded_at: '2026-11-21T19:04:25.795Z'
id: FACT-SUBJECT-UI-ASK-MODEL-QUESTION-METHOD
type: fact
---
Which HTTP requests of the built-in UI may send a question to the Ask box's model: the box's own POST, under the same-origin check, and not a GET of /ask or of the overview with a question in its query. It does not cover the commands, which answer either way, nor the guards a posted question then passes.