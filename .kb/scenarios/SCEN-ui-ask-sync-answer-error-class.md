---
title: The sync answer names the error class, how long the sync has failed and the latest attempts
status: active
origin:
  kind: agent
  ref: 'AB5 brief from the design session: bench fixes before the next run and the sync card''s error class (D-194, D-198, D-203, D-217), 2026-11-14'
  recorded_at: '2026-11-14T19:19:56.000Z'
id: SCEN-ui-ask-sync-answer-error-class
type: scenario
---
Given a connector whose sync keeps failing, one that failed once and one that works, when the owner asks why it fails, then the card says since when and for how many days it has been failing (less than a day on the day itself), how many attempts in a row failed and the latest error class named as the error class, the last good sync and the next try, and each of the latest attempts shows its time, that its figure is the error class, and the class (a sync that went through shows that), and a sync that works says nothing of failing; the card never holds an error's text.