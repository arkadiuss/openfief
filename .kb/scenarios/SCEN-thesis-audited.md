---
title: Given the real core, when the owner saves a draft and edits it, then the audit log holds exactly one ok
status: active
origin:
  kind: agent
  ref: 'Implement D-225 and D-229: opinions and draft theses (design session brief, 2026-11-21)'
  recorded_at: '2026-11-21T15:00:00.000Z'
id: SCEN-thesis-audited
type: scenario
---
Given the real core, when the owner saves a draft and edits it, then the audit log holds exactly one ok thesis_saved and one thesis_edited entry of the owner's session, naming the thesis and the version; and with the outbox gone neither call succeeds.