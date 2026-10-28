---
title: A model portfolio is stored only with valid targets whose weights sum to 100%
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:24.132Z'
id: SCEN-portfolio-model-validate
type: scenario
---
Given targets entered by the owner, when the model is validated, then each target must be an instrument (named ISIN/symbol, without currency) or an index (with its currency), named once, weighted above 0% and at most 100% in steps of 0.01%, at most 50 of them, summing to exactly 100%; an off model may have no targets, an on model must have some.