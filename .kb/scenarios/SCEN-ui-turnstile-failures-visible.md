---
title: A failed Turnstile check says why, on the card or in one message in the answer
status: active
origin:
  kind: agent
  ref: Ask box check failures made visible and specific, D-208, 2026-11-18
  recorded_at: '2026-11-18T15:15:42.840Z'
id: SCEN-ui-turnstile-failures-visible
type: scenario
---
Given a UI with Turnstile configured and a model on the Ask box, when a question comes with no token, with a token Cloudflare refuses, with one made for another host, or when the check can't be made, then each gets its own card (the refusal listing Cloudflare's error codes, the other host naming both hosts) and nothing is sent or counted; and in the browser, when the check cannot finish because Cloudflare reports an error, no token comes in 15 seconds, the script is blocked or the browser is unsupported, the answer area shows one message with the code or a fixed word and what to do, and the box is free to ask again.