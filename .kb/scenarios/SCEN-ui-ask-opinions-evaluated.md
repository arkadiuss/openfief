---
title: Given a model that passed the Ask box eval and no opinions set verdict, when it answers with opinions on, then the note says not evaluated; with opinions off it says evaluated
status: active
origin:
  kind: agent
  ref: Opinions eval set design answers, rounds 1 and 2 (design session, 2026-11-21)
  recorded_at: '2026-11-21T22:30:00.000Z'
id: SCEN-ui-ask-opinions-evaluated
type: scenario
---
Given a model with a passing verdict on the Ask box eval and no verdict of the opinions set, when it answers a question with opinions on, then the note names it as not evaluated; when it answers with opinions off it is evaluated; given an opinions-set verdict for the model, the opinions answer is evaluated. Assumes the committed result files hold a set field that tells the opinions set's runs from the eval's.