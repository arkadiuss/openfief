---
title: The pass is a signed cookie for one session and a cookie that does not verify is the same as none
status: active
origin:
  kind: agent
  ref: 'The design session''s amendment of D-208 (docs/DESIGN.md): a visitor who passes Turnstile isn''t checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
id: SCEN-ui-turnstile-pass-cookie
type: scenario
---
Given a visitor who passed the check, when the response is read, then it sets a cookie that is HttpOnly, Secure, SameSite=Strict, scoped to the Ask box's path and holds only a version, two times and a MAC; when the cookie is sent by another session, changed, truncated, made under another secret or malformed, the question is told a check is needed and no error is shown, and no log line holds the cookie, the session ID or the secret.