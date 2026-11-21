---
title: A valid pass skips the check but the core still counts the question and a command needs none
status: active
origin:
  kind: agent
  ref: 'The design session''s amendment of D-208 (docs/DESIGN.md): a visitor who passes Turnstile isn''t checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
id: SCEN-ui-turnstile-pass-skips-the-check
type: scenario
---
Given a visitor with a valid pass, when they ask a model question with no token, Cloudflare is not called and the core counts it against the session's and the address's limit and the day's budget, which can still refuse it; when they ask a command it is answered by the box with no check and no cookie.