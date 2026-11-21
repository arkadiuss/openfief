---
title: Given a saved draft, when the owner edits it twice, then it has three versions, oldest first, with the first
status: active
origin:
  kind: agent
  ref: 'Implement D-225 and D-229: opinions and draft theses (design session brief, 2026-11-21)'
  recorded_at: '2026-11-21T15:00:00.000Z'
id: SCEN-thesis-versions-kept
type: scenario
---
Given a saved draft, when the owner edits it twice, then it has three versions, oldest first, with the first text intact; when an edit names an older version it is refused and stores nothing; and the database refuses to update, delete or truncate a version.