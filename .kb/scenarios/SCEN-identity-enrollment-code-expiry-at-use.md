---
title: An invitation that expires while the viewer is enrolling creates no viewer
status: active
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #4 (2026-11-21)'
  recorded_at: '2026-11-21T18:39:00.063Z'
id: SCEN-identity-enrollment-code-expiry-at-use
type: scenario
---
Given a viewer's invitation that lasts an hour, when the viewer begins enrolling a minute before it expires and finishes two minutes later, then the registration is refused as forbidden, no viewer is created and the code stays unused; a code used a second before its expiry is spent as before.