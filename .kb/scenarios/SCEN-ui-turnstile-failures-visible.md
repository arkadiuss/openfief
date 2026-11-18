---
title: A failed Turnstile check says why on the card and in a status line
status: active
origin:
  kind: agent
  ref: Ask box check failures made visible and specific, D-208, 2026-11-18
  recorded_at: '2026-11-18T15:15:42.840Z'
id: SCEN-ui-turnstile-failures-visible
type: scenario
---
Given a UI with Turnstile configured and a model on the Ask box, when a question comes with no token, with a token Cloudflare refuses, with one made for another host, or when the check can't be made, then each gets its own card (the refusal listing Cloudflare's error codes, the other host naming both hosts) and nothing is sent or counted; and in the browser the status line under the field is a polite live region that says the check is running, is ready once a token arrives, and shows the widget's error code with what to do when it fails or when Cloudflare's script is blocked.