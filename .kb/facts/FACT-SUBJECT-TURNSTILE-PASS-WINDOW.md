---
title: Turnstile Pass Window
status: active
fact_kind: subject
subject_key: turnstile.pass_window
origin:
  kind: agent
  ref: 'The design session''s amendment of D-208 (docs/DESIGN.md): a visitor who passes Turnstile isn''t checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:18:42.747Z'
id: FACT-SUBJECT-TURNSTILE-PASS-WINDOW
type: fact
---
How long a visitor who passed the Turnstile check goes unchecked: the 30 minutes after their last model question, renewed by each model question that reaches the model, and the 2 hours from the check that end it however often it was renewed. It does not cover the cookie that carries it or the check it replaces.