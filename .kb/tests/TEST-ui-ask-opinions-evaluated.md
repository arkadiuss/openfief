---
title: 'An answer given with opinions on is evaluated only by the opinions set: a main-set pass does not count'
status: active
sourceFile: internal/views/ui/askevalnote_test.go
verification_scope: unit
origin:
  kind: agent
  ref: Opinions eval set design answers, rounds 1 and 2 (design session, 2026-11-21)
  recorded_at: '2026-11-21T22:30:00.000Z'
id: TEST-ui-ask-opinions-evaluated
type: test
---
A table test of noteAnswerer.answer: the same model with a pass of the Ask box eval, asked with opinions on, is not evaluated (with no opinions results, or with results of another model); with an opinions verdict it is; asked with opinions off it is evaluated by the main set only. A second test reads the result files: a run with set opinions never makes a model evaluated for the eval, and only a full run with a verdict makes one evaluated for opinions. It would pass falsely if the note read the main set for an opinion, which the first table catches.