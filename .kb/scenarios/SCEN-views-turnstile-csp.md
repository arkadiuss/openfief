---
title: The core's content security policy allows Turnstile's script and frame only when asked
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:41:42.014Z'
id: SCEN-views-turnstile-csp
type: scenario
---
Given the core started with and without --allow-turnstile, when it proxies a page or refuses a path then the policy is the strict one, or the same with script-src and frame-src allowing https://challenges.cloudflare.com and no wildcard, no unsafe keyword and no other host.