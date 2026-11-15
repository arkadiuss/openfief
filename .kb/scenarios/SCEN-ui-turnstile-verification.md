---
title: A question to the model passes only a check Cloudflare accepts
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T22:47:04.749Z'
id: SCEN-ui-turnstile-verification
type: scenario
---
Given a fake siteverify, when a question goes with a token it accepts then the model is asked once and the core counts the question after; with no token, a forged, used or other-host token, or when Cloudflare can't be reached, answers no JSON or refuses the secret, the box says a check comes first or that it couldn't run, with its commands, and neither the model nor the core's count is reached; the secret is read from its file on each call and nothing secret reaches a log; and the plugin's config sets the site key and URL together or neither.