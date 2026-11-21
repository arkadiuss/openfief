---
title: A follow-up carries the earlier turns, a command joins the thread, New conversation clears it and nothing is stored
status: active
origin:
  kind: agent
  ref: 'Implement D-224: the Ask box holds a conversation (design session brief, 2026-11-21); the author''s feedback that the Ask button must sit inline with the field'
  recorded_at: '2026-11-21T09:22:39.000Z'
id: SCEN-ui-ask-conversation
type: scenario
---
Given the Ask box with a model, when a question is answered and a follow-up is asked, then both turns are in the thread, the follow-up's POST body carries the first turn as its history and no URL holds either question; when a command is asked in the conversation the box answers it itself and the next question carries its turn; when New conversation is chosen the thread is empty and the next question carries no history; and at 360 and 390 px nothing was written to localStorage, sessionStorage or cookies and the page does not scroll sideways.