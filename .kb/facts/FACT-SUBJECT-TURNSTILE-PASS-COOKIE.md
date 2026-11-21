---
title: Turnstile Pass Cookie
status: active
fact_kind: subject
subject_key: turnstile.pass_cookie
origin:
  kind: agent
  ref: 'The design session''s amendment of D-208 (docs/DESIGN.md): a visitor who passes Turnstile isn''t checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:18:44.671Z'
id: FACT-SUBJECT-TURNSTILE-PASS-COOKIE
type: fact
---
The signed cookie that tells the Ask box a visitor's session passed the Turnstile check: what it holds, how it is signed and tied to a session, how it is scoped, and what happens when it fails to verify. It does not cover how long it lasts or the check it replaces.