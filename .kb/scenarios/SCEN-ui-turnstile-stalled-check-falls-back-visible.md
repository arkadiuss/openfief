---
title: A stalled invisible Turnstile check becomes a visible checkbox and the path is logged
status: active
origin:
  kind: agent
  ref: 'Design session: Turnstile never completes on Brave for Android, D-208 (pure UX), 2026-11-18'
  recorded_at: '2026-11-18T21:25:36.441Z'
id: SCEN-ui-turnstile-stalled-check-falls-back-visible
type: scenario
---
Given a UI with Turnstile configured and a model on the Ask box, when the widget's script loads but no token, no before-interactive callback and no error arrive for ten seconds, then the widget is removed and rendered again as the visible one with the same four callbacks, the status line says "The check didn't finish by itself. Complete the check below.", and once a token arrives it says Ready and the next question carries the token and the path visible-fallback; a check that finishes by itself, or one that asked for a click, is left alone and sends the path invisible; and the server logs the verdict with the path, keeping only the two fixed words (anything else is logged as unknown and never echoed).