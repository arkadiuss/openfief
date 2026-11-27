---
title: A viewer's questions to a model are limited by default and the owner's are not
status: active
origin:
  kind: agent
  ref: the owner's items for alpha.2, from the reviews of main 0bfeb62
  recorded_at: '2026-11-27T08:18:09.532Z'
id: SCEN-models-ask-viewer-question-limits
type: scenario
---
Given no question limit set for everyone, when a viewer's session asks 10 questions in an hour then the 11th is refused with the moment the hour's first question leaves the window, and the owner's session asks 50 without a refusal; the owner can lift or change the viewer's limits with --ask-viewer-questions-per-hour and --ask-viewer-questions-per-day, a limit set for everyone and a viewer's own both hold, the stricter of each window, and an hour above a day is refused at start.