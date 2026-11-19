---
title: 'Given a UI with Turnstile configured and a model on the Ask box, when a command is asked it is answered at once with no '
status: active
origin:
  kind: agent
  ref: 'Author feedback on demo.openfief.dev from Brave for Android (2026-11-19): a token reported ready but never sent, the check re-running, no progress, too much text; the design session approved the on-demand check and the single attribution note, D-208 and D-222 unchanged'
  recorded_at: '2026-11-19'
id: SCEN-ui-ask-turnstile-on-demand
type: scenario
---
Given a UI with Turnstile configured and a model on the Ask box, when a command is asked it is answered at once with no check; when a model question is asked it goes without a token, is told a check is needed by the HX-Trigger event openfief-check-needed and never sees the no-token card, the page shows Checking you're a person, runs the widget once, and asks the same question again with the token it made, and the answer comes; the form never holds a token, nothing is checked again after the answer, and the next model question makes one new check; a model question with no token that no script answers is still stopped by the server before anything is sent or counted.