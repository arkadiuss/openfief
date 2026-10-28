---
title: 'Conflict between design-decisions:D-26#1 and design-decisions:D-28#3: T-11: passkeys (D-26) bind to the origin''s domain, but D-28''s default access is localhost via port-forward, so passkeys registered there break behind a later Ingress hostname.'
status: closed
fact_kind: observation
text_ref: design-decisions:D-26#1; design-decisions:D-28#3
tags:
  - bootstrap
  - review:conflict
  - knowledge-source:design-decisions
origin:
  kind: agent
  recorded_at: '2026-10-27T20:13:04.504Z'
id: FACT-BOOTSTRAP-CONFLICT-0BF835EDD63962C4
type: fact
---
Tension T-11, resolved by the author in decision D-60 (docs/DESIGN.md): passkeys are kept per hostname. The first one is created on localhost through a one-time link the CLI prints, and a logged-in session issues a one-time link for each hostname the owner adds later.
