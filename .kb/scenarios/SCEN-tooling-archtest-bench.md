---
title: Only the bench imports goeval, and only its light packages
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:45.232Z'
id: SCEN-tooling-archtest-bench
type: scenario
---
Given fixture modules, when the architecture test runs, then the bench importing goeval's api, heuristic and openai packages passes, any other package or test importing goeval is a violation, the bench importing goeval's root or Gemini package is a violation, and the bench may import an adapter and the demo dataset where another tool may not.