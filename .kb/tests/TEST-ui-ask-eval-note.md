---
title: A model's answer shows its name, the disclaimer and the link to the eval, a command's answer shows none
sourceFile: internal/views/ui/askevalnote_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'AB9 brief: the Ask box evaluation article and the Ask box showing who answered, with the author''s D-222 (2026-11-15)'
  recorded_at: '2026-11-15T09:51:10.592Z'
status: active
id: TEST-ui-ask-eval-note
type: test
---
Through the page: an evaluated model, a model the eval has not covered and a refusal a model wrote show the model's name, the disclaimer and the link (and the not-evaluated mark where it applies); the commands' answers, with a model configured or not, show none of it. A table test of noteAnswerer (a model's card gets a note; a command's card and a spend guard's refusal card do not) and of which results make a model evaluated, against the committed result files. The browser check (scripts/ui-browser.mjs, task ui:browser) asserts the note and the link on a model's answer and its absence on a command's.