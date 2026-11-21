---
title: Issue, verify, slide, cap, refuse another session's, a tampered and an expired pass; a valid pass skips siteverify; limits still count; nothing is logged
status: active
origin:
  kind: agent
  ref: 'The design session''s amendment of D-208 (docs/DESIGN.md): a visitor who passes Turnstile isn''t checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
sourceFile: internal/views/ui/turnstilepass_test.go
verification_scope: integration
id: TEST-ui-turnstile-pass
type: test
---
Through the UI over its fake core and a fake siteverify (a TLS httptest server), with a clock the test moves: the check leaves a pass with its flags, prefix and nothing but a version, two times and a MAC; a valid pass skips siteverify, the fake is not called, and the core still counts each question and refuses past the session's and the address's limit and the budget; the window slides 20 minutes at a time to the cap of 2 hours from the check and is refused past it; an expired pass, one made for another session, one with a changed MAC, end, check time or version, an extra or a missing part, garbage and an over-long value, and one made under another secret, all fall back to the no-token card with the openfief-check-needed trigger and no error; a command needs none; a refused token or question earns or renews none; nothing of the cookie, the session ID or the secret is in the log. A false pass would be a fake that is called or a model that is not.