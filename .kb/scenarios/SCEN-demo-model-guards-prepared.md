---
title: The demo's model layer carries D-208's spend guards and a live demo cannot include it with a placeholder site key
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-15T00:17:03.778Z'
id: SCEN-demo-model-guards-prepared
type: scenario
---
Given the demo's model layer, when the demo is validated, then the core's flags set a positive daily budget, 10 questions an hour and 30 a day, 3 calls at once, the proxy whose address header the core believes and Turnstile's place in its policy, the built-in UI's config has the limits of one question and a site key, and its grant approves the host of the verification URL as an egress host and names the Secret ui--turnstile-secret; a flag that is missing or wrong, a budget that is no positive amount, a grant without the host or the Secret, and a live demo that includes the layer while the site key is still the placeholder each fail naming the problem.