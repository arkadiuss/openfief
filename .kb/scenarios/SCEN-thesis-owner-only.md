---
title: Given a core with an owner, a viewer, a demo session and a script token, when each calls SaveThesis,
status: active
origin:
  kind: agent
  ref: 'Implement D-225 and D-229: opinions and draft theses (design session brief, 2026-11-21)'
  recorded_at: '2026-11-21T15:00:00.000Z'
id: SCEN-thesis-owner-only
type: scenario
---
Given a core with an owner, a viewer, a demo session and a script token, when each calls SaveThesis, EditThesis, GetThesis and ListTheses, then the owner's calls work and every other caller is refused with permission_denied or unauthenticated, and no method of the read API returns a thesis.