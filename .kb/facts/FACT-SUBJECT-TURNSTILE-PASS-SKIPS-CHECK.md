---
title: Turnstile Pass Skips The Check
status: active
fact_kind: subject
subject_key: turnstile.pass_skips_check
origin:
  kind: agent
  ref: 'The design session''s amendment of D-208 (docs/DESIGN.md): a visitor who passes Turnstile isn''t checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:18:46.607Z'
id: FACT-SUBJECT-TURNSTILE-PASS-SKIPS-CHECK
type: fact
---
What a valid Turnstile pass changes for a model question and what it leaves alone: the question goes on with no token and no call to siteverify, while the core still counts it against the limits and the budget, and a fixed command never needs it.