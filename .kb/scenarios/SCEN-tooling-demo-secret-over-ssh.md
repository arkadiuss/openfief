---
title: SCEN-tooling-demo-secret-over-ssh
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T19:05:44.937Z'
id: SCEN-tooling-demo-secret-over-ssh
type: scenario
---
Given a fake ssh and a fake hcloud, when the model command runs with a key in the environment or piped in, then ssh is called once for root at the machine's address with a remote command that names openrouter--api-key and api-key and holds no value, the key reaches ssh on standard input exactly and appears in no argument, log or output, and a second run sends the new value; the Turnstile command does the same for ui--turnstile-secret and secret-key; without DEMO_SSH=keys, a value, a Hetzner token, a valid ENV or an existing machine, or with white space in the value, nothing is sent; and a failing ssh fails the command without claiming success.