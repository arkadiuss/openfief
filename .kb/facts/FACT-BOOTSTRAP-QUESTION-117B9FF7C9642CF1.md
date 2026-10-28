---
title: What is OpenFief's canonical hostname, given that passkeys are bound to it and M1's phone approvals need it reachable over HTTPS? (T-11, P-CC-29.)
status: closed
fact_kind: observation
text_ref: design-open-questions:OQ-24
tags:
  - bootstrap
  - knowledge-source:design-open-questions
  - review:open-question
origin:
  kind: agent
  recorded_at: '2026-10-27T20:13:11.271Z'
id: FACT-BOOTSTRAP-QUESTION-117B9FF7C9642CF1
type: fact
---
Open question OQ-24, answered by the author in decision D-60 (docs/DESIGN.md): OpenFief installs with no hostname, the first passkey is created on localhost, and a logged-in session issues a one-time link for a passkey on each hostname added later.
