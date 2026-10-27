---
title: 'Conflict between design-decisions:D-26#1 and design-decisions:D-28#3: T-11: passkeys (D-26) bind to the origin''s domain, but D-28''s default access is localhost via port-forward, so passkeys registered there break behind a later Ingress hostname.'
status: active
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
T-11: passkeys (D-26) bind to the origin's domain, but D-28's default access is localhost via port-forward, so passkeys registered there break behind a later Ingress hostname.

## Evidence

- OpenFief author decisions (DESIGN.md section 10) - D-26#1: People must log in with built-in passkeys (WebAuthn).
  > People log in with passkeys (WebAuthn), built in. No external identity provider.
- OpenFief author decisions (DESIGN.md section 10) - D-28#3: Default access to OpenFief must be through kubectl port-forward.
  > OpenFief ships no Ingress and no tunnel. The core and UI are ClusterIP services; the default way in is `kubectl port-forward`; anything else is the user's own Ingress, and the docs explain what it exposes.
