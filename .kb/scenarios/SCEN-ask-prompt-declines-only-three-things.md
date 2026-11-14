---
title: The Ask box declines a change, advice to buy or sell and a question outside the portfolio, and nothing else
status: active
origin:
  kind: agent
  ref: 'AB4 brief from the design session: softer decline wording, injection cases split into safety and answer, reasoning control and exact cost in the bench, 2026-11-14'
  recorded_at: '2026-11-14T16:54:51.212Z'
id: SCEN-ask-prompt-declines-only-three-things
type: scenario
---
Given the Ask box's system prompt, when its sentences are read, then exactly one names what is declined and it lists exactly three things (changing anything, advice on what to buy, sell or hold, a question that is not about the portfolio), no other sentence names one, the prompt tells the model to greet back and to answer every other question about the portfolio, why-questions included, from the cards, and a refusal is one friendly sentence.